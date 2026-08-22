package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBURL         string
	JWTSigningKey string
	LogLevel      string
	LogFormat     string
	AutoMigrate   bool
	RedisURL      string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	signingKey := os.Getenv("AUTH_JWT_SECRET")
	if len(signingKey) < 32 {
		return nil, fmt.Errorf("AUTH_JWT_SECRET must contain at least 32 bytes")
	}

	autoMigrate, _ := strconv.ParseBool(getEnv("AUTO_MIGRATE", "false"))

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBURL:         getEnv("DB_URL", "postgresql://postgres:password@localhost:5432/xlink?sslmode=disable"),
		JWTSigningKey: signingKey,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogFormat:     getEnv("LOG_FORMAT", "json"),
		AutoMigrate:   autoMigrate,
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
