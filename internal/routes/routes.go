package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(
	router *gin.Engine,
	urlHandler *handlers.URLHandler,
	authHandler *handlers.AuthHandler,
	healthHandler *handlers.HealthHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	authMiddleware gin.HandlerFunc,
) {
	// Prometheus metrics scrape endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health probes
	if healthHandler != nil {
		router.GET("/healthz", healthHandler.Healthz)
		router.GET("/livez", healthHandler.Livez)
		router.GET("/readyz", healthHandler.Readyz)
	}

	// Public short links
	router.GET("/:shortCode", urlHandler.Redirect)
	router.GET("/static/redirect", func(c *gin.Context) {
		c.Redirect(302, "https://google.com")
	})

	v1 := router.Group("/api/v1")

	if healthHandler != nil {
		v1.GET("/health", healthHandler.Healthz)
	}

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

	if analyticsHandler != nil {
		urls.GET("/:shortCode/analytics", analyticsHandler.GetAnalytics)
	}
}
