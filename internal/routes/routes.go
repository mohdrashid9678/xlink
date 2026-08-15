package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/handlers"
)

func RegisterRoutes(router *gin.Engine, urlHandler *handlers.URLHandler, authHandler *handlers.AuthHandler, authMiddleware gin.HandlerFunc) {
	// Public short links are separate from the versioned management API.
	router.GET("/:shortCode", urlHandler.Redirect)
	router.GET("/static/redirect", func(c *gin.Context) {
		c.Redirect(302, "https://google.com")
	})

	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.POST("/refresh", authHandler.Refresh)
	authRoutes.POST("/logout", authHandler.Logout)

	urls := v1.Group("/urls", authMiddleware)
	urls.POST("", urlHandler.Create)
	urls.GET("/:shortCode", urlHandler.Get)
	urls.PATCH("/:shortCode", urlHandler.Update)
	urls.DELETE("/:shortCode", urlHandler.Delete)
}
