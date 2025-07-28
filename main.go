package main

import (
	// db "backend-go/database"
	"backend-go/config"
	db "backend-go/database"
	registerhandler "backend-go/handlers"
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
	r.HandleFunc("/register", registerhandler.RegisterHandler).Methods("POST")

	log.Println("🚀 Server is running on port 8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), r))
}
