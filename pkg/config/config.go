package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBURL         string
	JWTSigningKey string
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	signingKey := os.Getenv("AUTH_JWT_SECRET")
	if len(signingKey) < 32 {
		return nil, fmt.Errorf("AUTH_JWT_SECRET must contain at least 32 bytes")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBURL:         getEnv("DB_URL", "postgresql://postgres:password@localhost:5432/xlink?sslmode=disable"),
		JWTSigningKey: signingKey,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
