package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}
}

func GetEnv(key, fallback string) string {
	fmt.Print(key)
	if value, exists := os.LookupEnv(key); exists {
		fmt.Println(value, exists)

		return value
	}
	return fallback
}
