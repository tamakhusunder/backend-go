package handlers

import (
	db "backend-go/database"
	usermodel "backend-go/models"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds usermodel.User
	payloadErr := json.NewDecoder(r.Body).Decode(&creds)
	if payloadErr != nil || creds.Email == "" || creds.Password == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	creds.Password = string(hashed)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.UserCollection.InsertOne(ctx, creds)
	if err != nil {
		http.Error(w, "User already exists or DB error", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
