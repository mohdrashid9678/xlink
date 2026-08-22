package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/cache"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"golang.org/x/sync/singleflight"
)

type URLService interface {
	Create(ctx context.Context, userID uuid.UUID, input models.CreateURLRequest) (*models.URL, error)
	Get(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error)
	GetPublic(ctx context.Context, shortCode string) (*models.URL, error)
	Update(ctx context.Context, userID uuid.UUID, shortCode string, input models.UpdateURLRequest) (*models.URL, error)
	Delete(ctx context.Context, userID uuid.UUID, shortCode string) error
	IncrementClickCount(ctx context.Context, shortCode string) error
}

type DefaultURLService struct {
	repository repository.URLRepository
	cache      cache.URLCache
	sf         singleflight.Group
}

func NewURLService(repository repository.URLRepository, cache cache.URLCache) *DefaultURLService {
	return &DefaultURLService{
		repository: repository,
		cache:      cache,
	}
}

func (s *DefaultURLService) Create(ctx context.Context, userID uuid.UUID, input models.CreateURLRequest) (*models.URL, error) {
	longURL, err := validateLongURL(input.LongURL)
	if err != nil {
		return nil, err
	}
	if err := validateExpiry(input.ExpiresAt); err != nil {
		return nil, err
	}

	var customAlias *string
	if input.CustomAlias != nil {
		alias, err := validateShortCode(*input.CustomAlias)
		if err != nil {
			return nil, err
		}
		customAlias = &alias
	}

	attempts := 1
	if customAlias == nil {
		attempts = 3
	}
	id := uuid.New()
	for attempt := 0; attempt < attempts; attempt++ {
		shortCode := ""
		if customAlias != nil {
			shortCode = *customAlias
		} else {
			shortCode = generateShortCode(id, uint64(attempt))
		}

		created, err := s.repository.Create(ctx, &models.URL{
			ID:          id,
			UserID:      userID,
			ShortCode:   shortCode,
			LongURL:     longURL,
			CustomAlias: customAlias,
			ExpiresAt:   input.ExpiresAt,
		})
		if err == nil {
			return created, nil
		}
		if errors.Is(err, repository.ErrConflict) {
			if customAlias != nil {
				return nil, ErrConflict
			}
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("generate a unique short code after %d attempts: %w", attempts, ErrConflict)
}

func (s *DefaultURLService) Get(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	shortCode, err := validateShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	url, err := s.repository.GetByShortCode(ctx, userID, shortCode)
	return url, mapRepositoryError(err)
}

func (s *DefaultURLService) GetPublic(ctx context.Context, shortCode string) (*models.URL, error) {
	shortCode, err := validateShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// 1. Fast-path: Check L2 Cache
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, shortCode)
		if err == nil && cached != nil {
			return cached, nil
		}
		if errors.Is(err, cache.ErrNotFoundCached) {
			return nil, ErrNotFound
		}
	}

	// 2. Cache Stampede & Penetration Prevention via singleflight
	val, err, _ := s.sf.Do(shortCode, func() (any, error) {
		// Double-check cache in case another flight populated it
		if s.cache != nil {
			if cached, err := s.cache.Get(ctx, shortCode); err == nil && cached != nil {
				return cached, nil
			}
			if errors.Is(err, cache.ErrNotFoundCached) {
				return nil, ErrNotFound
			}
		}

		url, err := s.repository.GetPublicByShortCode(ctx, shortCode)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				if s.cache != nil {
					_ = s.cache.SetNotFound(ctx, shortCode)
				}
				return nil, ErrNotFound
			}
			return nil, mapRepositoryError(err)
		}

		if s.cache != nil && url != nil {
			_ = s.cache.Set(ctx, url)
		}

		return url, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*models.URL), nil
}

func (s *DefaultURLService) Update(ctx context.Context, userID uuid.UUID, shortCode string, input models.UpdateURLRequest) (*models.URL, error) {
	shortCode, err := validateShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	if !input.HasUpdates() {
		return nil, fmt.Errorf("at least one editable field is required: %w", ErrValidation)
	}
	if input.LongURL != nil {
		value, err := validateLongURL(*input.LongURL)
		if err != nil {
			return nil, err
		}
		input.LongURL = &value
	}
	if input.CustomAlias != nil {
		value, err := validateShortCode(*input.CustomAlias)
		if err != nil {
			return nil, err
		}
		input.CustomAlias = &value
	}
	if err := validateExpiry(input.ExpiresAt); err != nil {
		return nil, err
	}

	updated, err := s.repository.Update(ctx, userID, shortCode, input)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, shortCode)
		if input.CustomAlias != nil && *input.CustomAlias != shortCode {
			_ = s.cache.Delete(ctx, *input.CustomAlias)
		}
	}

	return updated, nil
}

func (s *DefaultURLService) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	shortCode, err := validateShortCode(shortCode)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, userID, shortCode); err != nil {
		return mapRepositoryError(err)
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, shortCode)
	}

	return nil
}

func (s *DefaultURLService) IncrementClickCount(ctx context.Context, shortCode string) error {
	shortCode, err := validateShortCode(shortCode)
	if err != nil {
		return err
	}
	return mapRepositoryError(s.repository.IncrementClickCount(ctx, shortCode))
}
