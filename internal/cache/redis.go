package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache: key not found")

const (
	defaultPoolSize     = 100
	defaultMinIdleConns = 20
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 500 * time.Millisecond
	defaultWriteTimeout = 500 * time.Millisecond
)

type RedisConfig struct {
	URL          string
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	} else {
		opts.PoolSize = defaultPoolSize
	}

	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	} else {
		opts.MinIdleConns = defaultMinIdleConns
	}

	if cfg.DialTimeout > 0 {
		opts.DialTimeout = cfg.DialTimeout
	} else {
		opts.DialTimeout = defaultDialTimeout
	}

	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = cfg.ReadTimeout
	} else {
		opts.ReadTimeout = defaultReadTimeout
	}

	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = cfg.WriteTimeout
	} else {
		opts.WriteTimeout = defaultWriteTimeout
	}

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisClient{rdb: rdb}, nil
}

func (c *RedisClient) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}
	return nil
}

func (c *RedisClient) Close() error {
	return c.rdb.Close()
}

func (c *RedisClient) Raw() *redis.Client {
	return c.rdb
}
