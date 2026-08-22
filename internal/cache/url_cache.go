package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/mohdrashid9678/xlink/internal/models"
)

const (
	urlCacheKeyPrefix = "url:"
	defaultURLTTL     = 24 * time.Hour
	defaultJitterMax  = 30 * time.Minute
)

type URLCache interface {
	Get(ctx context.Context, shortCode string) (*models.URL, error)
	Set(ctx context.Context, url *models.URL) error
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

	return c.client.Set(ctx, urlCacheKeyPrefix+url.ShortCode, data, ttl)
}

func (c *RedisURLCache) Delete(ctx context.Context, shortCode string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Delete(ctx, urlCacheKeyPrefix+shortCode)
}
