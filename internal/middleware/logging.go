package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Duration("latency", latency),
			slog.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("bytes_out", c.Writer.Size()),
		}

		if rawQuery != "" {
			attrs = append(attrs, slog.String("query", rawQuery))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		ctx := c.Request.Context()
		switch {
		case status >= 500:
			log.LogAttrs(ctx, slog.LevelError, "HTTP request completed", attrs...)
		case status >= 400:
			log.LogAttrs(ctx, slog.LevelWarn, "HTTP request completed", attrs...)
		default:
			log.LogAttrs(ctx, slog.LevelInfo, "HTTP request completed", attrs...)
		}
	}
}
