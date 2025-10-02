package handlers

import (
	"backend-go/constants"
	contextkeys "backend-go/contextKeys"
	domainerrors "backend-go/internal/errors"
	"backend-go/internal/user/services"
	model "backend-go/models"
	userType "backend-go/type"
	"backend-go/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler interface {
	RegisterUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
	LogoutUser(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	GetSilentAccesToken(w http.ResponseWriter, r *http.Request)
}

type UserHandlerImpl struct {
	userService services.UserService
}

func NewUserHandler(s services.UserService) *UserHandlerImpl {
	return &UserHandlerImpl{
		userService: s,
	}
}

func (h *UserHandlerImpl) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var creds model.User
	//check payload
	payloadErr := json.NewDecoder(r.Body).Decode(&creds)
	if payloadErr != nil || creds.Email == "" || creds.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	creds.Password = string(hashed)
	defer cancel()

	_, err := h.userService.Register(ctx, creds)
	if err != nil {
		http.Error(w, "User already exists or DB error", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (h *UserHandlerImpl) LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds model.User
	w.Header().Set("Content-Type", "application/json")
	payloadErr := json.NewDecoder(r.Body).Decode(&creds)
	if payloadErr != nil || creds.Email == "" || creds.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientIp := utils.GetClientIP(r)

	userRes, err := h.userService.Login(ctx, creds.Email, creds.Password, clientIp)
	saveTokenInHttpCookie(w, userRes.AccessToken, userRes.RefreshToken)

	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrUserNotFound):
			http.Error(w, "user not found", http.StatusNotFound)
		case errors.Is(err, domainerrors.ErrInvalidCredentials):
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Login successful",
		"access_token": userRes.AccessToken,
		"user_id":      userRes.User.ID,
		"email":        userRes.User.Email,
	})
}

func (h *UserHandlerImpl) Profile(w http.ResponseWriter, r *http.Request) {
	userContent, ok := r.Context().Value(contextkeys.UserKey).(userType.UserContents)
	log.Printf("Claims in profile handler: %+v, ok: %v\n", userContent.Claims, ok)

	if !ok {
		http.Error(w, "Could not get user info", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userContent.Claims.UserID,
		"email":   userContent.Claims.Email,
	})
}

func (h *UserHandlerImpl) GetSilentAccesToken(w http.ResponseWriter, r *http.Request) {
	userContent, ok := r.Context().Value(contextkeys.UserKey).(userType.UserContents)

	if !ok {
		http.Error(w, "Could not get user info", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.userService.GetSilentAccessToken(context.Background(), userContent.Claims.UserID, userContent.Claims.Email)
	if err != nil || accessToken == "" {
		http.Error(w, "Could not get silent access token", http.StatusInternalServerError)
		return
	}

	saveAccesTokenInHttpCookie(w, accessToken)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func (h *UserHandlerImpl) LogoutUser(w http.ResponseWriter, r *http.Request) {
	userContent, ok := r.Context().Value(contextkeys.UserKey).(userType.UserContents)

	if !ok {
		http.Error(w, "Could not get user info", http.StatusUnauthorized)
		return
	}

	_, err := h.userService.Logout(context.Background(), userContent.Claims.UserID, userContent.AccessToken)
	if err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	clearTokenInHttpCookie(w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}

// internal functions
func saveTokenInHttpCookie(w http.ResponseWriter, accessToken string, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Expires:  time.Now().Add(constants.REFRESH_TOKEN_EXPIRATION),
	})
	saveAccesTokenInHttpCookie(w, accessToken)
}

func saveAccesTokenInHttpCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Expires:  time.Now().Add(constants.ACCESS_TOKEN_EXPIRATION),
	})
}

func clearTokenInHttpCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
}
