package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewRedisClientWithInvalidURL(t *testing.T) {
	_, err := NewRedisClient(RedisConfig{
		URL: "invalid://redis-url:notaport",
	})
	if err == nil {
		t.Fatal("expected error on invalid redis url, got nil")
	}
}

func TestNewRedisClientUnreachableHost(t *testing.T) {
	_, err := NewRedisClient(RedisConfig{
		URL:         "redis://127.0.0.1:59999/0",
		DialTimeout: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error on unreachable redis host, got nil")
	}
}

func TestRedisClientIntegration(t *testing.T) {
	// Attempt connection to local redis if available
	client, err := NewRedisClient(RedisConfig{
		URL:         "redis://localhost:6379/0",
		DialTimeout: 200 * time.Millisecond,
		ReadTimeout: 200 * time.Millisecond,
	})
	if err != nil {
		t.Skipf("skipping live redis integration test: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:xlink:key"
	val := "https://example.com/live"

	if err := client.Set(ctx, key, val, 10*time.Second); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if got != val {
		t.Fatalf("expected value %q, got %q", val, got)
	}

	if err := client.Delete(ctx, key); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	_, err = client.Get(ctx, key)
	if err != ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}
}
