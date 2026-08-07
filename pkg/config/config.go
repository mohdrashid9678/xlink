package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBUrl     string
	RedisAddr string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBUrl:     getEnv("DB_URL", "postgresql://postgres:password@localhost:5432/ticketrush?sslmode=disable"), // The default is not good at all, we will change it later
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
