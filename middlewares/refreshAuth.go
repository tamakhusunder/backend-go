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
		//validate the refreshToken
		isRefreshTokenVerified, err := isRefreshTokenVerified(r, UserRedisRepo)
		if err != nil || !isRefreshTokenVerified {
			http.Error(w, "Unauthorized - invalid refresh token", http.StatusUnauthorized)
			return
		}

		//validate the accessToken
		accessToken, errToken := utils.ExtractTokenFromHeader(r)
		if errToken != nil || accessToken == "" {
			log.Println("Error Token", errToken)
			http.Error(w, "Missing or invalid Access Token", http.StatusUnauthorized)
			return
		}

		accessTokenClaims, tokenErr := utils.VerifyAndParseJWTToken(accessToken)
		if tokenErr != nil {
			log.Println("Token from header:", accessTokenClaims, tokenErr)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		isBlacklisted, err := UserRedisRepo.IsBlacklistedAccessToken(r.Context(), accessTokenClaims.UserID, accessToken)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if isBlacklisted {
			log.Printf("Token %s for user %s is blacklisted", accessToken, accessTokenClaims.UserID)
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		userContents := userType.UserContents{
			Claims:      accessTokenClaims,
			AccessToken: accessToken,
		}
		ctx := context.WithValue(r.Context(), contextkeys.UserKey, userContents)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isRefreshTokenVerified(r *http.Request, UserRedisRepo redisRepository.UserRedisRepository) (bool, error) {
	refreshToken, errRefreshToken := r.Cookie("refresh_token")
	if errRefreshToken != nil {
		log.Println("Missing or invalid Refresh Token", errRefreshToken)
		return false, errRefreshToken
	}

	refreshTokenClaims, refreshTokenErr := utils.VerifyAndParseJWTToken(refreshToken.Value)
	if refreshTokenErr != nil {
		log.Println("Invalid token", refreshTokenErr)
		return false, refreshTokenErr
	}

	// verify token expiry and check ip address
	unixTimeSeconds := time.Now().Unix()
	if refreshTokenClaims.ExpiresAt == nil || refreshTokenClaims.ExpiresAt.Unix() < unixTimeSeconds {
		storedRdxToken, err := UserRedisRepo.GetToken(r.Context(), refreshTokenClaims.UserID)
		clientIp := utils.GetClientIP(r)
		if err != nil || storedRdxToken == nil {
			log.Printf("Failed to get user session from Redis: %v", err)
			return false, err
		} else if storedRdxToken.IPAddress != clientIp {
			log.Printf("IP address mismatch: token IP %s, request IP %s", storedRdxToken.IPAddress, clientIp)
			return false, nil
		}
	}

	return true, nil
}
