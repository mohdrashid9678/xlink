package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type mockURLService struct{}

func (m mockURLService) Create(ctx context.Context, userID uuid.UUID, input models.CreateURLRequest) (*models.URL, error) {
	return &models.URL{ID: uuid.New(), ShortCode: "mocked", LongURL: input.LongURL}, nil
}
func (m mockURLService) Get(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	return &models.URL{ID: uuid.New(), ShortCode: shortCode}, nil
}
func (m mockURLService) GetPublic(ctx context.Context, shortCode string) (*models.URL, error) {
	return &models.URL{ID: uuid.New(), ShortCode: shortCode, LongURL: "https://example.com"}, nil
}
func (m mockURLService) Update(ctx context.Context, userID uuid.UUID, shortCode string, input models.UpdateURLRequest) (*models.URL, error) {
	return &models.URL{ID: uuid.New(), ShortCode: shortCode}, nil
}
func (m mockURLService) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	return nil
}
func (m mockURLService) IncrementClickCount(ctx context.Context, shortCode string) error {
	return nil
}

type mockAuthService struct{}

func (m mockAuthService) Register(ctx context.Context, input models.RegisterRequest) (*models.AuthResponse, error) {
	return &models.AuthResponse{
		User: models.User{ID: uuid.New(), Email: input.Email, Name: input.Name},
		Tokens: models.TokenPair{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
		},
	}, nil
}
func (m mockAuthService) Login(ctx context.Context, input models.LoginRequest) (*models.AuthResponse, error) {
	return &models.AuthResponse{
		User: models.User{ID: uuid.New(), Email: input.Email},
		Tokens: models.TokenPair{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
		},
	}, nil
}
func (m mockAuthService) Refresh(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	return &models.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}, nil
}
func (m mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func TestRegisteredRoutesMatchExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	urlHandler := handlers.NewURLHandler(mockURLService{})
	authHandler := handlers.NewAuthHandler(mockAuthService{})
	healthHandler := handlers.NewHealthHandler(nil, nil)

	mockAuthMiddleware := func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, uuid.New())
		c.Next()
	}

	RegisterRoutes(router, urlHandler, authHandler, healthHandler, mockAuthMiddleware)

	tests := []struct {
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{"GET", "/metrics", "", http.StatusOK},
		{"GET", "/healthz", "", http.StatusOK},
		{"GET", "/livez", "", http.StatusOK},
		{"GET", "/readyz", "", http.StatusOK},
		{"GET", "/api/v1/health", "", http.StatusOK},
		{"POST", "/api/v1/auth/register", `{"email":"a@b.com","password":"Password123!","name":"Test"}`, http.StatusCreated},
		{"POST", "/api/v1/auth/login", `{"email":"a@b.com","password":"Password123!"}`, http.StatusOK},
		{"POST", "/api/v1/auth/refresh", `{"refresh_token":"abc"}`, http.StatusOK},
		{"POST", "/api/v1/auth/logout", `{"refresh_token":"abc"}`, http.StatusOK},
		{"POST", "/api/v1/urls", `{"long_url":"https://example.com"}`, http.StatusCreated},
		{"GET", "/api/v1/urls/test", "", http.StatusOK},
		{"PATCH", "/api/v1/urls/test", `{"long_url":"https://updated.com"}`, http.StatusOK},
		{"DELETE", "/api/v1/urls/test", "", http.StatusOK},
		{"GET", "/docs", "", http.StatusFound},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("%s %s expected status %d, got %d, body: %s", tt.method, tt.path, tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
