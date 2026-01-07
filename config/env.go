package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Loads environment variables on startup.
func init() {
	_ = godotenv.Load()
}

// GetEnv returns the value of the environment variable identified by key.
// If the variable is not set, the provided fallback value is returned.
func GetEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
