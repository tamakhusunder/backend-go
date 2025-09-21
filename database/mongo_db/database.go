package db

import (
	"backend-go/config"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitDB() (*mongo.Database, error) {
	var MongoURI = config.GetEnv("MONGO_URI", "mongodb://localhost:27018")
	var MongoDBName = config.GetEnv("DB_NAME", "backend_app_db")

	clientOptions := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	fmt.Println("✅ Connected to MongoDB")
	return client.Database(MongoDBName), nil
}
