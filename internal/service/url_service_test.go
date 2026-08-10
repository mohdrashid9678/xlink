package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
)

type stubURLRepository struct {
	create func(context.Context, *models.URL) (*models.URL, error)
	get    func(context.Context, string) (*models.URL, error)
	update func(context.Context, string, models.UpdateURLRequest) (*models.URL, error)
	delete func(context.Context, string) error
	incr   func(context.Context, string) error
}

func (r stubURLRepository) Create(ctx context.Context, url *models.URL) (*models.URL, error) {
	return r.create(ctx, url)
}

func (r stubURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	return r.get(ctx, shortCode)
}

func (r stubURLRepository) Update(ctx context.Context, shortCode string, update models.UpdateURLRequest) (*models.URL, error) {
	return r.update(ctx, shortCode, update)
}

func (r stubURLRepository) Delete(ctx context.Context, shortCode string) error {
	return r.delete(ctx, shortCode)
}

func (r stubURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	return r.incr(ctx, shortCode)
}

func TestCreateRejectsInvalidURLBeforeRepository(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		create: func(context.Context, *models.URL) (*models.URL, error) {
			t.Fatal("repository must not be called for invalid input")
			return nil, nil
		},
	})

	_, err := svc.Create(context.Background(), models.CreateURLRequest{LongURL: "not-a-url"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateUsesCustomAliasAsShortCode(t *testing.T) {
	alias := "docs"
	svc := NewURLService(stubURLRepository{
		create: func(_ context.Context, url *models.URL) (*models.URL, error) {
			if url.ShortCode != alias || url.CustomAlias == nil || *url.CustomAlias != alias {
				t.Fatalf("unexpected URL model: %#v", url)
			}
			url.CreatedAt = time.Now().UTC()
			return url, nil
		},
	})

	created, err := svc.Create(context.Background(), models.CreateURLRequest{
		LongURL:     "https://example.com/docs",
		CustomAlias: &alias,
	})
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("expected service to generate a UUID")
	}
}

func TestCreateRetriesWithDifferentHashBasedShortCode(t *testing.T) {
	var shortCodes []string
	svc := NewURLService(stubURLRepository{
		create: func(_ context.Context, url *models.URL) (*models.URL, error) {
			shortCodes = append(shortCodes, url.ShortCode)
			if len(shortCodes) == 1 {
				return nil, repository.ErrConflict
			}
			return url, nil
		},
	})

	_, err := svc.Create(context.Background(), models.CreateURLRequest{LongURL: "https://example.com"})
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}
	if len(shortCodes) != 2 {
		t.Fatalf("expected two create attempts, got %d", len(shortCodes))
	}
	if shortCodes[0] == shortCodes[1] {
		t.Fatalf("expected a different short code after collision, got %q", shortCodes[0])
	}
	for _, shortCode := range shortCodes {
		if len(shortCode) != shortCodeLength {
			t.Errorf("short code %q has length %d, want %d", shortCode, len(shortCode), shortCodeLength)
		}
	}
}

func TestGetMapsMissingURLToNotFound(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		get: func(context.Context, string) (*models.URL, error) {
			return nil, repository.ErrNotFound
		},
	})

	_, err := svc.Get(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateRequiresAnEditableField(t *testing.T) {
	svc := NewURLService(stubURLRepository{})

	_, err := svc.Update(context.Background(), "example", models.UpdateURLRequest{})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestIncrementClickCountMapsMissingURLToNotFound(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		incr: func(context.Context, string) error {
			return repository.ErrNotFound
		},
	})

	err := svc.IncrementClickCount(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
