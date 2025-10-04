package middleware

// TODO:
// - make a protected route that requires JWT authentication and public route
// - create a JWT middleware that checks for the token in the request header

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	contextkeys "backend-go/contextKeys"
	redisRepository "backend-go/internal/user/repository/redis"
	userType "backend-go/type"
	"backend-go/utils"
)

// AuthMiddleware checks for a valid JWT token in the request header

func AuthMiddleware(next http.Handler, UserRedisRepo redisRepository.UserRedisRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, errToken := utils.ExtractTokenFromHeader(r)
		if errToken != nil || accessToken == "" {
			fmt.Println("Error Token", errToken)
			http.Error(w, "Missing or invalid Access Token", http.StatusUnauthorized)
			return
		}

		//verify the token
		claims, tokenErr := utils.VerifyAndParseJWTToken(accessToken)
		if tokenErr != nil {
			fmt.Println("Token from header:", claims, tokenErr)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check blacklist access token
		isBlacklisted, err := UserRedisRepo.IsBlacklistedAccessToken(r.Context(), claims.UserID, accessToken)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if isBlacklisted {
			log.Printf("Token %s for user %s is blacklisted", accessToken, claims.UserID)
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		// verify token expiry
		unixTimeSeconds := time.Now().Unix()
		if claims.ExpiresAt == nil || claims.ExpiresAt.Unix() < unixTimeSeconds {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		//verify the token ip address with the stored ip address in redis
		storedRdxToken, err := UserRedisRepo.GetToken(r.Context(), claims.UserID)
		if err != nil {
			log.Printf("Failed to get user session from Redis: %v", err)
			http.Error(w, "internastoredDatal error", http.StatusInternalServerError)
			return
		}
		if storedRdxToken == nil {
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		clientIp := utils.GetClientIP(r)
		if storedRdxToken.IPAddress != clientIp {
			log.Printf("IP address mismatch: token IP %s, request IP %s", storedRdxToken.IPAddress, clientIp)
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		// TODO: Extract user from the DB using claims.UserID if needed

		// Attach user info into context
		userContents := userType.UserContents{
			Claims:      claims,
			AccessToken: accessToken,
		}
		ctx := context.WithValue(r.Context(), contextkeys.UserKey, userContents)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
