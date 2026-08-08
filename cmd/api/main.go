package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/database"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/mohdrashid9678/xlink/internal/routes"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/mohdrashid9678/xlink/pkg/config"
)

func main() {
	r := gin.Default()

	cfg := config.LoadConfig()
	db, err := database.NewPostgres(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	urlRepository := repository.NewPostgresURLRepository(db.Pool)
	urlService := service.NewURLService(urlRepository)
	urlHandler := handlers.NewURLHandler(urlService)
	routes.RegisterRoutes(r, urlHandler)

	// Start server on port 8080 (default)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
