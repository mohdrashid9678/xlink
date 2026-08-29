package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockURLRepoForAnalytics struct {
	mock.Mock
}

func (m *mockURLRepoForAnalytics) Create(ctx context.Context, url *models.URL) (*models.URL, error) {
	panic("unexpected")
}
func (m *mockURLRepoForAnalytics) GetByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	args := m.Called(ctx, userID, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.URL), args.Error(1)
}
func (m *mockURLRepoForAnalytics) GetPublicByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	panic("unexpected")
}
func (m *mockURLRepoForAnalytics) Update(ctx context.Context, userID uuid.UUID, shortCode string, input models.UpdateURLRequest) (*models.URL, error) {
	panic("unexpected")
}
func (m *mockURLRepoForAnalytics) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	panic("unexpected")
}
func (m *mockURLRepoForAnalytics) IncrementClickCount(ctx context.Context, shortCode string) error {
	panic("unexpected")
}

type mockAnalyticsRepo struct {
	mock.Mock
}

func (m *mockAnalyticsRepo) BatchInsertClicks(ctx context.Context, clicks []models.ClickEvent) error {
	args := m.Called(ctx, clicks)
	return args.Error(0)
}

func (m *mockAnalyticsRepo) GetAnalyticsSummary(ctx context.Context, urlID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error) {
	args := m.Called(ctx, urlID, shortCode, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalyticsSummary), args.Error(1)
}

func TestAnalyticsService_GetURLAnalytics_Success(t *testing.T) {
	mockURLRepo := new(mockURLRepoForAnalytics)
	mockAnalyticsRepo := new(mockAnalyticsRepo)
	svc := NewAnalyticsService(mockURLRepo, mockAnalyticsRepo)

	userID := uuid.New()
	urlID := uuid.New()
	shortCode := "validCode123"

	mockURL := &models.URL{
		ID:        urlID,
		UserID:    userID,
		ShortCode: shortCode,
		LongURL:   "https://example.com",
	}

	expectedSummary := &models.AnalyticsSummary{
		ShortCode:   shortCode,
		TotalClicks: 42,
		From:        time.Now().AddDate(0, 0, -30),
		To:          time.Now(),
	}

	query := models.AnalyticsQuery{Interval: "day"}

	mockURLRepo.On("GetByShortCode", mock.Anything, userID, shortCode).Return(mockURL, nil)
	mockAnalyticsRepo.On("GetAnalyticsSummary", mock.Anything, urlID, shortCode, query).Return(expectedSummary, nil)

	summary, err := svc.GetURLAnalytics(context.Background(), userID, shortCode, query)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), summary.TotalClicks)
	mockURLRepo.AssertExpectations(t)
	mockAnalyticsRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetURLAnalytics_NotFound(t *testing.T) {
	mockURLRepo := new(mockURLRepoForAnalytics)
	mockAnalyticsRepo := new(mockAnalyticsRepo)
	svc := NewAnalyticsService(mockURLRepo, mockAnalyticsRepo)

	userID := uuid.New()
	shortCode := "notFoundCode"

	mockURLRepo.On("GetByShortCode", mock.Anything, userID, shortCode).Return(nil, repository.ErrNotFound)

	summary, err := svc.GetURLAnalytics(context.Background(), userID, shortCode, models.AnalyticsQuery{})

	assert.Nil(t, summary)
	assert.ErrorIs(t, err, ErrNotFound)
}
