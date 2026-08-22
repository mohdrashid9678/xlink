package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/pkg/logger"
)

func TestRequestIDMiddlewareSetsHeaderAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())

	var capturedReqID string
	router.GET("/test", func(c *gin.Context) {
		capturedReqID = logger.GetRequestID(c.Request.Context())
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(rec, req)

	respHeader := rec.Header().Get(RequestIDHeader)
	if respHeader == "" {
		t.Fatal("expected X-Request-ID header in response")
	}
	if capturedReqID != respHeader {
		t.Fatalf("expected context request ID %q to match response header %q", capturedReqID, respHeader)
	}
}

func TestStructuredLoggerLogsRequestDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buf := &bytes.Buffer{}
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(RequestID(), StructuredLogger(log))
	router.GET("/test-log", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-log?query=1", nil)
	req.Header.Set(RequestIDHeader, "custom-id-999")
	router.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output JSON: %v, raw: %s", err, buf.String())
	}

	if entry["request_id"] != "custom-id-999" {
		t.Errorf("expected request_id 'custom-id-999', got %v", entry["request_id"])
	}
	if entry["method"] != "GET" {
		t.Errorf("expected method 'GET', got %v", entry["method"])
	}
	if entry["path"] != "/test-log" {
		t.Errorf("expected path '/test-log', got %v", entry["path"])
	}
	if entry["query"] != "query=1" {
		t.Errorf("expected query 'query=1', got %v", entry["query"])
	}
}

func TestRecoveryMiddlewareCatchesPanicAndLogsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buf := &bytes.Buffer{}
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(RequestID(), Recovery(log))
	router.GET("/panic", func(c *gin.Context) {
		panic("database crashed")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log output: %v", err)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("expected error level log on panic, got %v", entry["level"])
	}
}
