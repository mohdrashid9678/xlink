package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/pkg/logger"
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

				reqID := logger.GetRequestID(c.Request.Context())
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.ProblemDetails{
					Type:      "urn:xlink:error:internal_server_error",
					Title:     "Internal Server Error",
					Status:    http.StatusInternalServerError,
					Detail:    "An unexpected error occurred.",
					Code:      "internal_server_error",
					Instance:  c.Request.URL.Path,
					RequestID: reqID,
				})
			}
		}()
		c.Next()
	}
}
