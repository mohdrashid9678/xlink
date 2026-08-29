package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnalyticsService struct {
	mock.Mock
}

func (m *mockAnalyticsService) GetURLAnalytics(ctx context.Context, userID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error) {
	args := m.Called(ctx, userID, shortCode, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalyticsSummary), args.Error(1)
}

func TestAnalyticsHandler_GetAnalytics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAnalyticsService)
	handler := NewAnalyticsHandler(mockSvc)

	userID := uuid.New()
	shortCode := "valid123"

	expectedSummary := &models.AnalyticsSummary{
		ShortCode:      shortCode,
		TotalClicks:    150,
		UniqueVisitors: 80,
		From:           time.Now().AddDate(0, 0, -7),
		To:             time.Now(),
		Countries: []models.DistributionItem{
			{Name: "IN", Count: 100, Percentage: 66.67},
			{Name: "US", Count: 50, Percentage: 33.33},
		},
		Devices: []models.DistributionItem{
			{Name: "mobile", Count: 100, Percentage: 66.67},
			{Name: "desktop", Count: 50, Percentage: 33.33},
		},
	}

	mockSvc.On("GetURLAnalytics", mock.Anything, userID, shortCode, mock.Anything).Return(expectedSummary, nil)

	router := gin.New()
	router.GET("/api/v1/urls/:shortCode/analytics", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, userID)
		handler.GetAnalytics(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/urls/"+shortCode+"/analytics?interval=day", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data models.AnalyticsSummary `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(150), resp.Data.TotalClicks)
	assert.Equal(t, int64(80), resp.Data.UniqueVisitors)
	assert.Equal(t, 2, len(resp.Data.Countries))
}

func TestAnalyticsHandler_GetAnalytics_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAnalyticsService)
	handler := NewAnalyticsHandler(mockSvc)

	router := gin.New()
	router.GET("/api/v1/urls/:shortCode/analytics", handler.GetAnalytics)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/urls/test/analytics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAnalyticsHandler_GetAnalytics_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockAnalyticsService)
	handler := NewAnalyticsHandler(mockSvc)

	userID := uuid.New()
	shortCode := "notfound123"

	mockSvc.On("GetURLAnalytics", mock.Anything, userID, shortCode, mock.Anything).Return(nil, service.ErrNotFound)

	router := gin.New()
	router.GET("/api/v1/urls/:shortCode/analytics", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, userID)
		handler.GetAnalytics(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/urls/"+shortCode+"/analytics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
