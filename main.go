package main

import (
	// db "backend-go/database"
	"backend-go/config"
	db "backend-go/database"
	handlers "backend-go/handlers"
	middleware "backend-go/middlewares"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// import (
// 	"log"

func main() {
	config.LoadEnv() // Load environment variables
	var PORT = config.GetEnv("PORT", "8080")

	db.InitDB() // Initialize MongoDB connection

	r := mux.NewRouter()
	//user
	r.HandleFunc("/api/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")

	//TODO
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")

	r.Handle("/api/profile", middleware.AuthMiddleware(http.HandlerFunc(handlers.ProfileHandler))).Methods("GET")
	r.HandleFunc("/refresh", handlers.RefreshHandler)

	log.Println("🚀 Server is running on port 8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), r))
}
