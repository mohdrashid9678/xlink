package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/pkg/tracer"
)

func TestOpenTelemetryTracingMiddlewareSetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, shutdown, err := tracer.InitTracer(context.Background(), tracer.Config{
		ServiceName:   "test-api",
		SamplingRatio: 1.0,
	})
	if err != nil {
		t.Fatalf("failed to init tracer: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	router := gin.New()
	router.Use(OpenTelemetryTracing("test-api"), TraceResponseHeader())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	traceID := rec.Header().Get("X-Trace-ID")
	if traceID == "" {
		t.Fatal("expected X-Trace-ID header to be present in response")
	}
}
