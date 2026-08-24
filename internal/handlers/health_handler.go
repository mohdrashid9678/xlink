package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/cache"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *cache.RedisClient
}

func NewHealthHandler(db *pgxpool.Pool, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "xlink",
	})
}

func (h *HealthHandler) Livez(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ALIVE",
	})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	isReady := true

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "DOWN"
			isReady = false
		} else {
			checks["database"] = "UP"
		}
	} else {
		checks["database"] = "N/A"
	}

	if h.redis != nil && h.redis.Raw() != nil {
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = "DEGRADED"
		} else {
			checks["redis"] = "UP"
		}
	} else {
		checks["redis"] = "DISABLED"
	}

	statusCode := http.StatusOK
	statusText := "READY"
	if !isReady {
		statusCode = http.StatusServiceUnavailable
		statusText = "NOT_READY"
	}

	c.JSON(statusCode, gin.H{
		"status":    statusText,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	})
}
