package cache

import (
	"context"
	"log/slog"
)

// NewURLCacheStack builds and initializes the complete multi-tier caching stack:
// L1 (Ristretto TinyLFU In-Memory) + L2 (Redis Distributed Cache) + Redis Pub/Sub Coordination.
// It returns a unified URLCache and a cleanup function to gracefully release resources on shutdown.
func NewURLCacheStack(ctx context.Context, redisURL string, log *slog.Logger) (URLCache, func()) {
	var (
		l1          L1Cache
		l2          *RedisURLCache
		coordinator CacheCoordinator
		cleanups    []func()
	)

	// 1. Initialize L1 In-Memory Cache
	if l1Cache, err := NewRistrettoL1Cache(); err != nil {
		log.Warn("L1 in-memory cache unavailable", slog.Any("error", err))
	} else {
		l1 = l1Cache
		cleanups = append(cleanups, l1Cache.Close)
		log.Info("L1 in-memory TinyLFU cache initialized")
	}

	// 2. Initialize L2 Redis & Pub/Sub Coordination
	if redisClient, err := NewRedisClient(RedisConfig{URL: redisURL}); err != nil {
		log.Warn("Redis connection unavailable, operating in database-only mode", slog.Any("error", err))
	} else {
		cleanups = append(cleanups, func() { _ = redisClient.Close() })
		l2 = NewRedisURLCache(redisClient)

		if l1 != nil {
			coord := NewRedisPubSubCoordinator(redisClient, l1, log)
			coord.Start(ctx)
			cleanups = append(cleanups, func() { _ = coord.Close() })
			coordinator = coord
		}
		log.Info("Multi-tier cache (L1 + L2 + Pub/Sub Coordination) initialized")
	}

	closeFunc := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	return NewMultiTierURLCache(l1, l2, coordinator), closeFunc
}
