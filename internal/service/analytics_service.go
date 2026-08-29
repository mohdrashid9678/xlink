package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type AnalyticsService interface {
	GetURLAnalytics(ctx context.Context, userID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error)
}

type DefaultAnalyticsService struct {
	urlRepo       repository.URLRepository
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(urlRepo repository.URLRepository, analyticsRepo repository.AnalyticsRepository) *DefaultAnalyticsService {
	return &DefaultAnalyticsService{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
	}
}

func (s *DefaultAnalyticsService) GetURLAnalytics(ctx context.Context, userID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error) {
	ctx, span := otel.Tracer("xlink-service").Start(ctx, "service.GetURLAnalytics",
		trace.WithAttributes(
			attribute.String("url.short_code", shortCode),
			attribute.String("user.id", userID.String()),
		),
	)
	defer span.End()

	cleanCode, err := validateShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 1. Authorize: Ensure the short code exists and belongs to the authenticated user
	url, err := s.urlRepo.GetByShortCode(ctx, userID, cleanCode)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	// 2. Fetch aggregated analytics
	return s.analyticsRepo.GetAnalyticsSummary(ctx, url.ID, url.ShortCode, query)
}
