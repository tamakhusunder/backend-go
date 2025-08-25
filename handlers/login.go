package handlers

import (
	"backend-go/constants"
	db "backend-go/database"
	usermodel "backend-go/models"
	"backend-go/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds usermodel.User
	payloadErr := json.NewDecoder(r.Body).Decode(&creds)
	if payloadErr != nil || creds.Email == "" || creds.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	var user usermodel.User
	err := db.UserCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	fmt.Print(creds)
	fmt.Print(user)

	// Compare hashed password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	accessToken, errAcessToken := utils.GenerateAccessToken(user.ID, user.Email)
	refreshToken, errRefreshToken := utils.GenerateRefreshToken(user.ID, user.Email)
	if errAcessToken != nil || errRefreshToken != nil {
		fmt.Print("No username found")
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Expires:  time.Now().Add(constants.REFRESH_TOKEN_EXPIRATION),
	})

	fmt.Println("Token generated successfully:", accessToken)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Login successful",
		"access_token": accessToken,
		"user_id":      user.ID,
		"email":        user.Email,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
