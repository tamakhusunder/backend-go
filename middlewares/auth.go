package middleware

// TODO:
// - make a protected route that requires JWT authentication and public route
// - create a JWT middleware that checks for the token in the request header

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	contextkeys "backend-go/contextKeys"
	redisRepository "backend-go/internal/user/repository/redis"
	userType "backend-go/type"
	"backend-go/utils"
)

// AuthMiddleware checks for a valid JWT token in the request header

func AuthMiddleware(next http.Handler, UserRedisRepo redisRepository.UserRedisRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, errToken := extractTokenFromHeader(r)
		if errToken != nil || token == "" {
			fmt.Println("Error Token", errToken)
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		//verify the token
		claims, tokenErr := utils.VerifyAndParseJWTToken(token)

		if tokenErr != nil {
			fmt.Println("Token from header:", claims, tokenErr)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check blacklist first
		isBlacklisted, err := UserRedisRepo.IsBlacklistedAccessToken(r.Context(), claims.UserID, token)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if isBlacklisted {
			log.Printf("Token %s for user %s is blacklisted", token, claims.UserID)
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		userContents := userType.UserContents{
			Claims:      claims,
			AccessToken: token,
		}

		//TODO : check the validation time of refresh and access token

		// TODO: Extract user from the DB using claims.UserID if needed

		// Attach user info into context
		ctx := context.WithValue(r.Context(), contextkeys.UserKey, userContents)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Extract token from Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	// Check for Bearer token format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
