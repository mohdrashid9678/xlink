package cache

import (
	"context"

	"github.com/mohdrashid9678/xlink/internal/models"
)

type MultiTierURLCache struct {
	l1          L1Cache
	l2          *RedisURLCache
	coordinator CacheCoordinator
}

func NewMultiTierURLCache(l1 L1Cache, l2 *RedisURLCache, coordinator CacheCoordinator) *MultiTierURLCache {
	return &MultiTierURLCache{
		l1:          l1,
		l2:          l2,
		coordinator: coordinator,
	}
}

func (c *MultiTierURLCache) Get(ctx context.Context, shortCode string) (*models.URL, error) {
	// 1. Fast-path: Check L1 Local Memory Cache (<50ns)
	if c.l1 != nil {
		if url, found := c.l1.Get(ctx, shortCode); found && url != nil {
			return url, nil
		}
	}

	// 2. Fallback: Check L2 Distributed Redis Cache (~0.8ms)
	if c.l2 != nil {
		url, err := c.l2.Get(ctx, shortCode)
		if err != nil {
			return nil, err
		}
		if url != nil {
			// Populate L1 cache for subsequent ultra-fast reads
			if c.l1 != nil {
				c.l1.Set(ctx, url, defaultL1TTL)
			}
			return url, nil
		}
	}

	return nil, ErrCacheMiss
}

func (c *MultiTierURLCache) Set(ctx context.Context, url *models.URL) error {
	if url == nil {
		return nil
	}

	if c.l1 != nil {
		c.l1.Set(ctx, url, defaultL1TTL)
	}

	if c.l2 != nil {
		return c.l2.Set(ctx, url)
	}

	return nil
}

func (c *MultiTierURLCache) SetNotFound(ctx context.Context, shortCode string) error {
	if c.l1 != nil {
		c.l1.Delete(ctx, shortCode)
	}

	if c.l2 != nil {
		return c.l2.SetNotFound(ctx, shortCode)
	}

	return nil
}

func (c *MultiTierURLCache) Delete(ctx context.Context, shortCode string) error {
	if c.l1 != nil {
		c.l1.Delete(ctx, shortCode)
	}

	var l2Err error
	if c.l2 != nil {
		l2Err = c.l2.Delete(ctx, shortCode)
	}

	if c.coordinator != nil {
		_ = c.coordinator.PublishInvalidation(ctx, shortCode)
	}

	return l2Err
}
