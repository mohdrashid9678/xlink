package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type stubURLService struct {
	get    func(context.Context, uuid.UUID, string) (*models.URL, error)
	public func(context.Context, string) (*models.URL, error)
	incr   func(context.Context, string) error
}

func (s stubURLService) Create(context.Context, uuid.UUID, models.CreateURLRequest) (*models.URL, error) {
	panic("unexpected Create call")
}

func (s stubURLService) Get(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	return s.get(ctx, userID, shortCode)
}

func (s stubURLService) GetPublic(ctx context.Context, shortCode string) (*models.URL, error) {
	return s.public(ctx, shortCode)
}

func (s stubURLService) Update(context.Context, uuid.UUID, string, models.UpdateURLRequest) (*models.URL, error) {
	panic("unexpected Update call")
}

func (s stubURLService) Delete(context.Context, uuid.UUID, string) error {
	panic("unexpected Delete call")
}

func (s stubURLService) IncrementClickCount(ctx context.Context, shortCode string) error {
	return s.incr(ctx, shortCode)
}

func TestRedirectTracksClickWithoutWaitingForCounterUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trackingStarted := make(chan struct{})
	releaseTracking := make(chan struct{})
	trackingComplete := make(chan struct{})

	handler := NewURLHandler(stubURLService{
		public: func(_ context.Context, shortCode string) (*models.URL, error) {
			if shortCode != "docs" {
				t.Fatalf("expected short code docs, got %q", shortCode)
			}
			return &models.URL{
				ID:        uuid.New(),
				ShortCode: shortCode,
				LongURL:   "https://example.com/docs",
			}, nil
		},
		incr: func(_ context.Context, shortCode string) error {
			if shortCode != "docs" {
				t.Errorf("expected tracked short code docs, got %q", shortCode)
			}
			close(trackingStarted)
			<-releaseTracking
			close(trackingComplete)
			return nil
		},
	})

	router := gin.New()
	router.GET("/:shortCode", handler.Redirect)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}
	if location := response.Header().Get("Location"); location != "https://example.com/docs" {
		t.Fatalf("expected redirect to destination, got %q", location)
	}
	select {
	case <-trackingStarted:
	case <-time.After(time.Second):
		t.Fatal("expected asynchronous click tracking to start")
	}
	close(releaseTracking)
	select {
	case <-trackingComplete:
	case <-time.After(time.Second):
		t.Fatal("expected asynchronous click tracking to finish")
	}
}

func TestGetDoesNotTrackClick(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewURLHandler(stubURLService{
		get: func(_ context.Context, _ uuid.UUID, shortCode string) (*models.URL, error) {
			return &models.URL{ID: uuid.New(), ShortCode: shortCode}, nil
		},
		incr: func(context.Context, string) error {
			t.Fatal("management GET must not increment click count")
			return nil
		},
	})

	router := gin.New()
	router.GET("/api/v1/urls/:shortCode", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, uuid.New())
		handler.Get(c)
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/urls/docs", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
