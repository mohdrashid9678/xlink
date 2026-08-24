package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandlerHealthzAndLivez(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler(nil, nil)

	router := gin.New()
	router.GET("/healthz", handler.Healthz)
	router.GET("/livez", handler.Livez)
	router.GET("/readyz", handler.Readyz)

	// Test /healthz
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /healthz status 200, got %d", rec.Code)
	}

	// Test /livez
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/livez", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /livez status 200, got %d", rec.Code)
	}

	// Test /readyz
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /readyz status 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json body: %v", err)
	}
	if body["status"] != "READY" {
		t.Fatalf("expected status READY, got %v", body["status"])
	}
}
