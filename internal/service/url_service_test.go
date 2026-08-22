package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/cache"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
)

type stubURLRepository struct {
	create func(context.Context, *models.URL) (*models.URL, error)
	get    func(context.Context, uuid.UUID, string) (*models.URL, error)
	public func(context.Context, string) (*models.URL, error)
	update func(context.Context, uuid.UUID, string, models.UpdateURLRequest) (*models.URL, error)
	delete func(context.Context, uuid.UUID, string) error
	incr   func(context.Context, string) error
}

func (r stubURLRepository) Create(ctx context.Context, url *models.URL) (*models.URL, error) {
	return r.create(ctx, url)
}
func (r stubURLRepository) GetByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*models.URL, error) {
	return r.get(ctx, userID, shortCode)
}
func (r stubURLRepository) GetPublicByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	return r.public(ctx, shortCode)
}
func (r stubURLRepository) Update(ctx context.Context, userID uuid.UUID, shortCode string, update models.UpdateURLRequest) (*models.URL, error) {
	return r.update(ctx, userID, shortCode, update)
}
func (r stubURLRepository) Delete(ctx context.Context, userID uuid.UUID, shortCode string) error {
	return r.delete(ctx, userID, shortCode)
}
func (r stubURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	return r.incr(ctx, shortCode)
}

type stubURLCache struct {
	get         func(ctx context.Context, shortCode string) (*models.URL, error)
	set         func(ctx context.Context, url *models.URL) error
	setNotFound func(ctx context.Context, shortCode string) error
	delete      func(ctx context.Context, shortCode string) error
}

func (c stubURLCache) Get(ctx context.Context, shortCode string) (*models.URL, error) {
	if c.get != nil {
		return c.get(ctx, shortCode)
	}
	return nil, cache.ErrCacheMiss
}
func (c stubURLCache) Set(ctx context.Context, url *models.URL) error {
	if c.set != nil {
		return c.set(ctx, url)
	}
	return nil
}
func (c stubURLCache) SetNotFound(ctx context.Context, shortCode string) error {
	if c.setNotFound != nil {
		return c.setNotFound(ctx, shortCode)
	}
	return nil
}
func (c stubURLCache) Delete(ctx context.Context, shortCode string) error {
	if c.delete != nil {
		return c.delete(ctx, shortCode)
	}
	return nil
}

func TestCreateRejectsInvalidURLBeforeRepository(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		create: func(context.Context, *models.URL) (*models.URL, error) {
			t.Fatal("repository must not be called for invalid input")
			return nil, nil
		},
	}, nil)

	_, err := svc.Create(context.Background(), uuid.New(), models.CreateURLRequest{LongURL: "not-a-url"})
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
	}, nil)

	created, err := svc.Create(context.Background(), uuid.New(), models.CreateURLRequest{
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
	}, nil)

	_, err := svc.Create(context.Background(), uuid.New(), models.CreateURLRequest{LongURL: "https://example.com"})
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
		get: func(context.Context, uuid.UUID, string) (*models.URL, error) {
			return nil, repository.ErrNotFound
		},
	}, nil)

	_, err := svc.Get(context.Background(), uuid.New(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetPublicReturnsCachedValueWithoutRepositoryCall(t *testing.T) {
	cachedURL := &models.URL{
		ID:        uuid.New(),
		ShortCode: "docs",
		LongURL:   "https://example.com/cached-docs",
	}

	svc := NewURLService(stubURLRepository{
		public: func(context.Context, string) (*models.URL, error) {
			t.Fatal("repository must NOT be called on cache hit")
			return nil, nil
		},
	}, stubURLCache{
		get: func(ctx context.Context, shortCode string) (*models.URL, error) {
			if shortCode == "docs" {
				return cachedURL, nil
			}
			return nil, cache.ErrCacheMiss
		},
	})

	url, err := svc.GetPublic(context.Background(), "docs")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if url.LongURL != cachedURL.LongURL {
		t.Fatalf("expected cached URL %q, got %q", cachedURL.LongURL, url.LongURL)
	}
}

func TestGetPublicPopulatesCacheOnMiss(t *testing.T) {
	dbURL := &models.URL{
		ID:        uuid.New(),
		ShortCode: "docs",
		LongURL:   "https://example.com/from-db",
	}

	var cacheSetCalled bool
	svc := NewURLService(stubURLRepository{
		public: func(ctx context.Context, shortCode string) (*models.URL, error) {
			return dbURL, nil
		},
	}, stubURLCache{
		get: func(ctx context.Context, shortCode string) (*models.URL, error) {
			return nil, cache.ErrCacheMiss
		},
		set: func(ctx context.Context, url *models.URL) error {
			cacheSetCalled = true
			if url.LongURL != dbURL.LongURL {
				t.Errorf("expected cached url to be %q, got %q", dbURL.LongURL, url.LongURL)
			}
			return nil
		},
	})

	url, err := svc.GetPublic(context.Background(), "docs")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if url.LongURL != dbURL.LongURL {
		t.Fatalf("expected URL %q, got %q", dbURL.LongURL, url.LongURL)
	}
	if !cacheSetCalled {
		t.Fatal("expected cache.Set to be called on cache miss")
	}
}

func TestGetPublicReturnsNotFoundFromNegativeCache(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		public: func(context.Context, string) (*models.URL, error) {
			t.Fatal("repository must NOT be called when negative cache is hit")
			return nil, nil
		},
	}, stubURLCache{
		get: func(ctx context.Context, shortCode string) (*models.URL, error) {
			return nil, cache.ErrNotFoundCached
		},
	})

	_, err := svc.GetPublic(context.Background(), "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound from negative cache, got %v", err)
	}
}

func TestSingleFlightCoalescesConcurrentRequests(t *testing.T) {
	dbURL := &models.URL{
		ID:        uuid.New(),
		ShortCode: "viral-link",
		LongURL:   "https://example.com/viral",
	}

	var repoCalls int64
	svc := NewURLService(stubURLRepository{
		public: func(ctx context.Context, shortCode string) (*models.URL, error) {
			atomic.AddInt64(&repoCalls, 1)
			time.Sleep(30 * time.Millisecond) // Simulate DB query latency
			return dbURL, nil
		},
	}, stubURLCache{
		get: func(ctx context.Context, shortCode string) (*models.URL, error) {
			return nil, cache.ErrCacheMiss
		},
		set: func(ctx context.Context, url *models.URL) error {
			return nil
		},
	})

	const concurrentGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(concurrentGoroutines)

	for i := 0; i < concurrentGoroutines; i++ {
		go func() {
			defer wg.Done()
			url, err := svc.GetPublic(context.Background(), "viral-link")
			if err != nil {
				t.Errorf("expected success, got error: %v", err)
			}
			if url.LongURL != dbURL.LongURL {
				t.Errorf("expected %q, got %q", dbURL.LongURL, url.LongURL)
			}
		}()
	}

	wg.Wait()

	if calls := atomic.LoadInt64(&repoCalls); calls != 1 {
		t.Fatalf("expected exactly 1 repository call through singleflight, but got %d", calls)
	}
}

func TestUpdateInvalidatesCache(t *testing.T) {
	var deletedKey string
	newURL := "https://example.com/updated"
	svc := NewURLService(stubURLRepository{
		update: func(ctx context.Context, u uuid.UUID, s string, r models.UpdateURLRequest) (*models.URL, error) {
			return &models.URL{ShortCode: s, LongURL: *r.LongURL}, nil
		},
	}, stubURLCache{
		delete: func(ctx context.Context, shortCode string) error {
			deletedKey = shortCode
			return nil
		},
	})

	_, err := svc.Update(context.Background(), uuid.New(), "docs", models.UpdateURLRequest{
		LongURL: &newURL,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deletedKey != "docs" {
		t.Fatalf("expected deleted cache key 'docs', got %q", deletedKey)
	}
}

func TestDeleteInvalidatesCache(t *testing.T) {
	var deletedKey string
	svc := NewURLService(stubURLRepository{
		delete: func(ctx context.Context, u uuid.UUID, s string) error {
			return nil
		},
	}, stubURLCache{
		delete: func(ctx context.Context, shortCode string) error {
			deletedKey = shortCode
			return nil
		},
	})

	err := svc.Delete(context.Background(), uuid.New(), "docs")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deletedKey != "docs" {
		t.Fatalf("expected deleted cache key 'docs', got %q", deletedKey)
	}
}

func TestUpdateRequiresAnEditableField(t *testing.T) {
	svc := NewURLService(stubURLRepository{}, nil)

	_, err := svc.Update(context.Background(), uuid.New(), "example", models.UpdateURLRequest{})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestIncrementClickCountMapsMissingURLToNotFound(t *testing.T) {
	svc := NewURLService(stubURLRepository{
		incr: func(context.Context, string) error {
			return repository.ErrNotFound
		},
	}, nil)

	err := svc.IncrementClickCount(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
