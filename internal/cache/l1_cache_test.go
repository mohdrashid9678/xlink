package cache

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
)

func TestRistrettoL1CacheGetSetDelete(t *testing.T) {
	l1, err := NewRistrettoL1Cache()
	if err != nil {
		t.Fatalf("failed to create l1 cache: %v", err)
	}
	defer l1.Close()

	ctx := context.Background()
	url := &models.URL{
		ID:        uuid.New(),
		ShortCode: "github",
		LongURL:   "https://github.com",
	}

	l1.Set(ctx, url, time.Minute)
	// Ristretto sets are asynchronous through internal ring buffers; wait slightly
	time.Sleep(10 * time.Millisecond)

	got, found := l1.Get(ctx, "github")
	if !found || got == nil {
		t.Fatal("expected key to be found in L1 cache")
	}
	if got.LongURL != url.LongURL {
		t.Fatalf("expected %q, got %q", url.LongURL, got.LongURL)
	}

	l1.Delete(ctx, "github")
	time.Sleep(10 * time.Millisecond)

	_, found = l1.Get(ctx, "github")
	if found {
		t.Fatal("expected key to be deleted from L1 cache")
	}
}
