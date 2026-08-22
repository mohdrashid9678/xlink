package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				log.ErrorContext(c.Request.Context(), "server panic recovered",
					slog.Any("panic", r),
					slog.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "internal_server_error",
						"message": "An unexpected error occurred.",
					},
				})
			}
		}()
		c.Next()
	}
}
