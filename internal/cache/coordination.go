package cache

import (
	"context"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
)

const InvalidationChannel = "xlink:cache:invalidate"

type CacheCoordinator interface {
	PublishInvalidation(ctx context.Context, shortCode string) error
	Start(ctx context.Context)
	Close() error
}

type RedisPubSubCoordinator struct {
	client *RedisClient
	l1     L1Cache
	logger *slog.Logger
	pubsub *redis.PubSub
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRedisPubSubCoordinator(client *RedisClient, l1 L1Cache, logger *slog.Logger) *RedisPubSubCoordinator {
	return &RedisPubSubCoordinator{
		client: client,
		l1:     l1,
		logger: logger,
	}
}

func (c *RedisPubSubCoordinator) PublishInvalidation(ctx context.Context, shortCode string) error {
	if c == nil || c.client == nil || shortCode == "" {
		return nil
	}
	return c.client.Raw().Publish(ctx, InvalidationChannel, shortCode).Err()
}

func (c *RedisPubSubCoordinator) Start(ctx context.Context) {
	if c == nil || c.client == nil || c.l1 == nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.pubsub = c.client.Raw().Subscribe(ctx, InvalidationChannel)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ch := c.pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg != nil && msg.Payload != "" {
					c.l1.Delete(ctx, msg.Payload)
				}
			}
		}
	}()
}

func (c *RedisPubSubCoordinator) Close() error {
	if c == nil {
		return nil
	}
	if c.cancel != nil {
		c.cancel()
	}
	var err error
	if c.pubsub != nil {
		err = c.pubsub.Close()
	}
	c.wg.Wait()
	return err
}
