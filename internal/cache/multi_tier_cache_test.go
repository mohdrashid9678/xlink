package cache

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type stubL1Cache struct {
	get    func(ctx context.Context, shortCode string) (*models.URL, bool)
	set    func(ctx context.Context, url *models.URL, ttl time.Duration)
	delete func(ctx context.Context, shortCode string)
}

func (s *stubL1Cache) Get(ctx context.Context, shortCode string) (*models.URL, bool) {
	if s.get != nil {
		return s.get(ctx, shortCode)
	}
	return nil, false
}
func (s *stubL1Cache) Set(ctx context.Context, url *models.URL, ttl time.Duration) {
	if s.set != nil {
		s.set(ctx, url, ttl)
	}
}
func (s *stubL1Cache) Delete(ctx context.Context, shortCode string) {
	if s.delete != nil {
		s.delete(ctx, shortCode)
	}
}
func (s *stubL1Cache) Close() {}

type stubCoordinator struct {
	published string
}

func (c *stubCoordinator) PublishInvalidation(ctx context.Context, shortCode string) error {
	c.published = shortCode
	return nil
}
func (c *stubCoordinator) Start(ctx context.Context) {}
func (c *stubCoordinator) Close() error               { return nil }

func TestMultiTierCacheReturnsL1HitWithoutHittingL2(t *testing.T) {
	cachedURL := &models.URL{
		ID:        uuid.New(),
		ShortCode: "gh",
		LongURL:   "https://github.com",
	}

	l1 := &stubL1Cache{
		get: func(ctx context.Context, shortCode string) (*models.URL, bool) {
			if shortCode == "gh" {
				return cachedURL, true
			}
			return nil, false
		},
	}

	multi := NewMultiTierURLCache(l1, nil, nil)
	got, err := multi.Get(context.Background(), "gh")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.LongURL != cachedURL.LongURL {
		t.Fatalf("expected %q, got %q", cachedURL.LongURL, got.LongURL)
	}
}

func TestMultiTierCacheEvictsL1AndPublishesOnDelete(t *testing.T) {
	var l1Deleted string
	coord := &stubCoordinator{}

	l1 := &stubL1Cache{
		delete: func(ctx context.Context, shortCode string) {
			l1Deleted = shortCode
		},
	}

	multi := NewMultiTierURLCache(l1, nil, coord)
	err := multi.Delete(context.Background(), "gh")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if l1Deleted != "gh" {
		t.Fatalf("expected L1 to delete 'gh', got %q", l1Deleted)
	}
	if coord.published != "gh" {
		t.Fatalf("expected coordinator to broadcast 'gh', got %q", coord.published)
	}
}
