package utils

import (
	"backend-go/config"
	"backend-go/constants"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var JWT_SECRET_KEY = []byte(config.GetEnv("JWT_SECRET", "your_secret_key"))

func generateJWTToken(userID string, email string, timeDuration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-go",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateAccessToken(userID string, email string) (string, error) {
	token, err := generateJWTToken(userID, email, constants.ACCESS_TOKEN_EXPIRATION) // 30 minutes
	return token, err
}

func GenerateRefreshToken(userID string, email string) (string, error) {
	token, err := generateJWTToken(userID, email, constants.REFRESH_TOKEN_EXPIRATION) //24 hours
	return token, err
}

func VerifyAndParseJWTToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWT_SECRET_KEY, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
