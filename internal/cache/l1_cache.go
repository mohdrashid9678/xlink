package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mohdrashid9678/xlink/internal/models"
)

const (
	defaultL1MaxCost     = 100000          // Max 100,000 URLs in RAM (~25MB)
	defaultL1NumCounters = 1000000         // 10x counters for TinyLFU accuracy
	defaultL1TTL         = 5 * time.Minute // 5-minute maximum L1 TTL for safety
)

type L1Cache interface {
	Get(ctx context.Context, shortCode string) (*models.URL, bool)
	Set(ctx context.Context, url *models.URL, ttl time.Duration)
	Delete(ctx context.Context, shortCode string)
	Close()
}

type RistrettoL1Cache struct {
	cache *ristretto.Cache[string, *models.URL]
}

func NewRistrettoL1Cache() (*RistrettoL1Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, *models.URL]{
		NumCounters: defaultL1NumCounters,
		MaxCost:     defaultL1MaxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("create ristretto l1 cache: %w", err)
	}

	return &RistrettoL1Cache{cache: cache}, nil
}

func (c *RistrettoL1Cache) Get(ctx context.Context, shortCode string) (*models.URL, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}
	val, found := c.cache.Get(shortCode)
	if !found || val == nil {
		return nil, false
	}
	return val, true
}

func (c *RistrettoL1Cache) Set(ctx context.Context, url *models.URL, ttl time.Duration) {
	if c == nil || c.cache == nil || url == nil {
		return
	}
	if ttl <= 0 || ttl > defaultL1TTL {
		ttl = defaultL1TTL
	}
	c.cache.SetWithTTL(url.ShortCode, url, 1, ttl)
}

func (c *RistrettoL1Cache) Delete(ctx context.Context, shortCode string) {
	if c == nil || c.cache == nil {
		return
	}
	c.cache.Del(shortCode)
}

func (c *RistrettoL1Cache) Close() {
	if c != nil && c.cache != nil {
		c.cache.Close()
	}
}
