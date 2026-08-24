package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// OpenTelemetryTracing instruments HTTP routes with OpenTelemetry spans and W3C trace propagation.
func OpenTelemetryTracing(serviceName string) gin.HandlerFunc {
	if serviceName == "" {
		serviceName = "xlink-api"
	}
	return otelgin.Middleware(serviceName)
}

// TraceResponseHeader exposes the active OpenTelemetry TraceID in the X-Trace-ID response header.
func TraceResponseHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			c.Header("X-Trace-ID", span.SpanContext().TraceID().String())
		}
		c.Next()
	}
}
