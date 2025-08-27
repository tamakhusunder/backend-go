package handlers

import (
	"backend-go/utils"
	"fmt"
	"net/http"
)

// RefreshHandler issues a new access token if the refresh token is valid
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get refresh token from HttpOnly cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Refresh token missing", http.StatusUnauthorized)
		return
	}

	// 2. Validate refresh token
	claims, err := utils.VerifyAndParseJWTToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	fmt.Printf("Claims in refresh handler: %+v\n", claims)

	// TODO: Extract user from the JWT token and generate a new access token
	var accessToken string
	// accessToken, err = utils.GenerateAccessToken(claims.UserID, claims.Email)

	// 3. Issue a new short-lived access token
	// accessToken, err := utils.GenerateAccessToken(claims["user_id"].(string))
	// if err != nil {
	// 	http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
	// 	return
	// }

	// 4. Send back the new access token in JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"access_token": "` + accessToken + `"}`))
}
