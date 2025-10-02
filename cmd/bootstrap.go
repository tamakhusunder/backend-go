package main

import (
	"backend-go/config"
	db "backend-go/database/mongo_db"
	"backend-go/database/redisx"
	uApp "backend-go/internal/user/app"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run() {
	loadConfig()
}

func loadConfig() {
	PORT := config.GetEnv("PORT", "8080")
	config.LoadEnv() // Load environment variables

	mongoDB, err := InitializeMongoDB()
	if err != nil {
		log.Fatal("❌ DB init failed: ", err)
	}

	redisDB, err := InitializeRedis()
	if err != nil {
		log.Fatal("❌ Redis init failed: ", err)
	}

	r := RegisterWebAppRouter(PORT, mongoDB, redisDB)
	InitializeWebServer(r, PORT)
}

func InitializeMongoDB() (*mongo.Database, error) {
	return db.InitDB()
}

func InitializeRedis() (*redisx.Client, error) {
	return redisx.InitRedis()
}

func RegisterWebAppRouter(port string, mongoDB *mongo.Database, redisDB *redisx.Client) *mux.Router {
	r := mux.NewRouter()

	userApp, err := uApp.NewApp(mongoDB, redisDB)
	if err != nil {
		log.Fatal("failed to initialize user app:", err)
	}
	userApp.RegisterRoutes(r.PathPrefix("/api/user").Subrouter())

	return r
}

func InitializeWebServer(r *mux.Router, port string) {
	log.Println("🚀 Server is running on port:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
