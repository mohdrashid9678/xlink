package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/database"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/mohdrashid9678/xlink/internal/routes"
	"github.com/mohdrashid9678/xlink/internal/server"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/mohdrashid9678/xlink/pkg/config"
	"github.com/mohdrashid9678/xlink/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("application terminated with error: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	appLogger := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})

	db, err := database.NewPostgres(cfg.DBURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	urlRepository := repository.NewPostgresURLRepository(db.Pool)
	userRepository := repository.NewPostgresUserRepository(db.Pool)
	refreshTokenRepository := repository.NewPostgresRefreshTokenRepository(db.Pool)

	jwtManager := auth.NewJWTManager(cfg.JWTSigningKey)
	urlService := service.NewURLService(urlRepository)
	authService := service.NewAuthService(userRepository, refreshTokenRepository, jwtManager)

	urlHandler := handlers.NewURLHandler(urlService)
	authHandler := handlers.NewAuthHandler(authService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.StructuredLogger(appLogger),
		middleware.Recovery(appLogger),
	)

	routes.RegisterRoutes(router, urlHandler, authHandler, middleware.RequireAuth(jwtManager))

	srv := server.New(router, server.Config{
		Port: cfg.Port,
	})

	return srv.Serve(ctx)
}
