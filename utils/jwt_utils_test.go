package utils_test

import (
	"backend-go/utils"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJWTToken_valid(t *testing.T) {
	token, err := utils.GenerateAccessToken("user123", "test@gmail.com")
	assert.NoError(t, err, "GenerateAccessToken failed")
	assert.NotEmpty(t, token, "GenerateAccessToken returned empty token")

	// userID, err := utils.ValidateJWT(token)
	// assert.NoError(t, err)
	// assert.Equal(t, "user123", userID)
}
func TestGenerateRefreshToken_valid(t *testing.T) {
	token, err := utils.GenerateRefreshToken("user456", "refresh@gmail.com")
	assert.NoError(t, err, "GenerateRefreshToken failed")
	assert.NotEmpty(t, token, "GenerateRefreshToken returned empty token")
}

func TestVerifyAndParseJWTToken_ValidToken(t *testing.T) {
	userID := "user789"
	email := "valid@gmail.com"
	token, err := utils.GenerateAccessToken(userID, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := utils.VerifyAndParseJWTToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestVerifyAndParseJWTToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"
	claims, err := utils.VerifyAndParseJWTToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestExtractTokenFromHeader_Valid(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken123")
	token, err := utils.ExtractTokenFromHeader(req)
	assert.NoError(t, err)
	assert.Equal(t, "sometoken123", token)
}

func TestExtractTokenFromHeader_MissingHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	token, err := utils.ExtractTokenFromHeader(req)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestExtractTokenFromHeader_InvalidFormat(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Token sometoken123")
	token, err := utils.ExtractTokenFromHeader(req)
	assert.Error(t, err)
	assert.Empty(t, token)
}
