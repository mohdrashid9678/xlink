package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/mohdrashid9678/xlink/pkg/metrics"
)

func TestPrometheusMetricsMiddlewareRecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	router := gin.New()
	router.Use(PrometheusMetrics(m))
	router.GET("/urls/:shortCode", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/urls/docs", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	count := testutil.ToFloat64(m.RequestsTotal.WithLabelValues(http.MethodGet, "/urls/:shortCode", "200"))
	if count != 1 {
		t.Fatalf("expected requests_total count 1, got %f", count)
	}
}
