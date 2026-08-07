package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Load Config
	// cfg := config.LoadConfig()

	// Database Connection
	// dbService, err := database.NewPostgres(cfg.DBUrl)
	// if err != nil {
	// 	log.Fatalf("Database connection failed: %v", err)
	// }
	// defer dbService.Close()

	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Start server on port 8080 (default)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
