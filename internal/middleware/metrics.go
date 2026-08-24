package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/pkg/metrics"
)

// PrometheusMetrics captures RED metrics for every HTTP request.
// It uses route template normalization (c.FullPath()) to strictly prevent cardinality explosion.
func PrometheusMetrics(m *metrics.Metrics) gin.HandlerFunc {
	if m == nil {
		m = metrics.DefaultMetrics
	}

	return func(c *gin.Context) {
		start := time.Now()

		handler := c.FullPath()
		if handler == "" {
			handler = "unmatched"
		}

		m.RequestsInFlight.WithLabelValues(handler).Inc()
		defer m.RequestsInFlight.WithLabelValues(handler).Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		m.RequestsTotal.WithLabelValues(c.Request.Method, handler, status).Inc()
		m.RequestDuration.WithLabelValues(c.Request.Method, handler, status).Observe(duration)
	}
}
