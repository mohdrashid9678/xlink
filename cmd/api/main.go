package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/database"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/mohdrashid9678/xlink/internal/routes"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/mohdrashid9678/xlink/pkg/config"
)

func main() {
	r := gin.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration failed: %v", err)
	}
	db, err := database.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	urlRepository := repository.NewPostgresURLRepository(db.Pool)
	urlService := service.NewURLService(urlRepository)
	urlHandler := handlers.NewURLHandler(urlService)
	userRepository := repository.NewPostgresUserRepository(db.Pool)
	refreshTokenRepository := repository.NewPostgresRefreshTokenRepository(db.Pool)
	jwtManager := auth.NewJWTManager(cfg.JWTSigningKey)
	authHandler := handlers.NewAuthHandler(service.NewAuthService(userRepository, refreshTokenRepository, jwtManager))
	routes.RegisterRoutes(r, urlHandler, authHandler, middleware.RequireAuth(jwtManager))

	// Start server on port 8080 (default)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
