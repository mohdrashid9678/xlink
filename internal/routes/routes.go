package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/handlers"
)

func RegisterRoutes(router *gin.Engine, urlHandler *handlers.URLHandler) {
	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	urls := v1.Group("/urls")
	urls.POST("", urlHandler.Create)
	urls.GET("/:shortCode", urlHandler.Get)
	urls.PATCH("/:shortCode", urlHandler.Update)
	urls.DELETE("/:shortCode", urlHandler.Delete)
}
