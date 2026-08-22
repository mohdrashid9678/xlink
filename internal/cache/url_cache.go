package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/mohdrashid9678/xlink/internal/models"
)

var ErrNotFoundCached = errors.New("cache: key marked as not found")

const (
	urlCacheKeyPrefix      = "url:"
	notFoundCacheKeyPrefix = "url:nf:"
	defaultURLTTL          = 24 * time.Hour
	defaultNotFoundTTL     = 60 * time.Second
	defaultJitterMax       = 30 * time.Minute
)

type URLCache interface {
	Get(ctx context.Context, shortCode string) (*models.URL, error)
	Set(ctx context.Context, url *models.URL) error
	SetNotFound(ctx context.Context, shortCode string) error
	Delete(ctx context.Context, shortCode string) error
}

type RedisURLCache struct {
	client *RedisClient
	ttl    time.Duration
}

func NewRedisURLCache(client *RedisClient) *RedisURLCache {
	return &RedisURLCache{
		client: client,
		ttl:    defaultURLTTL,
	}
}

func (c *RedisURLCache) Get(ctx context.Context, shortCode string) (*models.URL, error) {
	if c == nil || c.client == nil {
		return nil, ErrCacheMiss
	}

	// Check if key is cached as not found (cache penetration protection)
	_, nfErr := c.client.Get(ctx, notFoundCacheKeyPrefix+shortCode)
	if nfErr == nil {
		return nil, ErrNotFoundCached
	}

	raw, err := c.client.Get(ctx, urlCacheKeyPrefix+shortCode)
	if err != nil {
		return nil, err
	}

	var u models.URL
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		return nil, fmt.Errorf("unmarshal cached url: %w", err)
	}

	return &u, nil
}

func (c *RedisURLCache) Set(ctx context.Context, url *models.URL) error {
	if c == nil || c.client == nil || url == nil {
		return nil
	}

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("marshal url for cache: %w", err)
	}

	jitter := time.Duration(rand.Int63n(int64(defaultJitterMax)))
	ttl := c.ttl + jitter

	if url.ExpiresAt != nil {
		timeUntilExpiry := time.Until(*url.ExpiresAt)
		if timeUntilExpiry > 0 && timeUntilExpiry < ttl {
			ttl = timeUntilExpiry
		}
	}

	_ = c.client.Delete(ctx, notFoundCacheKeyPrefix+url.ShortCode)
	return c.client.Set(ctx, urlCacheKeyPrefix+url.ShortCode, data, ttl)
}

func (c *RedisURLCache) SetNotFound(ctx context.Context, shortCode string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, notFoundCacheKeyPrefix+shortCode, "1", defaultNotFoundTTL)
}

func (c *RedisURLCache) Delete(ctx context.Context, shortCode string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Delete(ctx, urlCacheKeyPrefix+shortCode, notFoundCacheKeyPrefix+shortCode)
}
