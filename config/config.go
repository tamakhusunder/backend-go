package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func IsLocal() bool {
	var env = GetEnv("ENV", "local")
	return env == "local" || env == "dev"
}
