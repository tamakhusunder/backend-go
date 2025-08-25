package middleware

// TODO:
// - make a protected route that requires JWT authentication and public route
// - create a JWT middleware that checks for the token in the request header

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"backend-go/utils"
)

// AuthMiddleware checks for a valid JWT token in the request header

type contextKey string

const UserContextKey = contextKey("user")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, errToken := extractTokenFromHeader(r)
		if errToken != nil || token == "" {
			fmt.Println("Error Token", errToken)
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		claims, err := utils.VerifyJWTToken(token)

		if err != nil {
			fmt.Println("Token from header:", claims, err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// TODO: Extract user from the DB using claims.UserID if needed

		// Attach user info into context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
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
