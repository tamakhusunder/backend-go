// internal/middleware/refresh_auth.go
package middleware

import (
	contextkeys "backend-go/contextKeys"
	redisRepository "backend-go/internal/user/repository/redis"
	userType "backend-go/type"
	"backend-go/utils"
	"context"
	"log"
	"net/http"
	"time"
)

func RefreshAuthMiddleware(next http.Handler, UserRedisRepo redisRepository.UserRedisRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check if user is blacklisted
		accessToken, errAccessToken := r.Cookie("access_token")
		if errAccessToken != nil {
			log.Println("Missing or invalid Access Token", errAccessToken)
			http.Error(w, "Missing or invalid Access Token", http.StatusUnauthorized)
			return
		}
		isNotBlacklisted, err := checkIsUserBlacklist(r, UserRedisRepo, accessToken.Value)
		if err != nil || !isNotBlacklisted {
			http.Error(w, "Unauthorized - user is blacklisted", http.StatusUnauthorized)
			return
		}

		//validate the refreshToken
		refreshToken, errToken := utils.ExtractTokenFromHeader(r)
		if errToken != nil || refreshToken == "" {
			log.Println("Error Token", errToken)
			http.Error(w, "Missing or invalid Access Token", http.StatusUnauthorized)
			return
		}

		refreshTokenClaims, tokenErr := utils.VerifyAndParseJWTToken(refreshToken)
		if tokenErr != nil {
			log.Println("Token from header:", refreshTokenClaims, tokenErr)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		//verify token expiry and check ip address
		unixTimeSeconds := time.Now().Unix()
		if refreshTokenClaims.ExpiresAt == nil || refreshTokenClaims.ExpiresAt.Unix() < unixTimeSeconds {
			log.Println("Refresh Token expired")
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		storedRdxToken, err := UserRedisRepo.GetToken(r.Context(), refreshTokenClaims.UserID)
		clientIp := utils.GetClientIP(r)
		if err != nil || storedRdxToken == nil {
			log.Printf("Failed to get user session from Redis: %v", err)
			http.Error(w, "Unauthorized - invalid session", http.StatusUnauthorized)
			return
		} else if storedRdxToken.IPAddress != clientIp || storedRdxToken.RefreshToken != refreshToken {
			log.Printf("IP address mismatch: token IP %s, request IP %s", storedRdxToken.IPAddress, clientIp)
			http.Error(w, "Unauthorized - invalid session", http.StatusUnauthorized)
			return
		}

		userContents := userType.UserContents{
			Claims:      refreshTokenClaims,
			AccessToken: accessToken.Value,
		}
		ctx := context.WithValue(r.Context(), contextkeys.UserKey, userContents)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkIsUserBlacklist(r *http.Request, UserRedisRepo redisRepository.UserRedisRepository, accessToken string) (bool, error) {
	accessTokenClaims, tokenErr := utils.VerifyAndParseJWTToken(accessToken)
	if tokenErr != nil {
		log.Println("Invalid token", tokenErr)
		return false, tokenErr
	}

	isBlacklisted, err := UserRedisRepo.IsBlacklistedAccessToken(r.Context(), accessTokenClaims.UserID, accessToken)
	if err != nil || isBlacklisted {
		return false, err
	}

	return true, nil
}
