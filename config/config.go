package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found (using system env vars instead)")
	}
}

func GetEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// GetDatabaseURL returns the database URL from environment variables
func GetDatabaseURL() string {
	return GetEnv("DATABASE_URL", "")
}
