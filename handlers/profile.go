package handlers

import (
	middleware "backend-go/middlewares"
	"backend-go/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Get user info from context set by AuthMiddleware
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	fmt.Printf("Claims in profile handler: %+v, ok: %v\n", claims, ok)
	if !ok {
		http.Error(w, "Could not get user info", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": claims.UserID,
		"email":   claims.Email,
	})
}
