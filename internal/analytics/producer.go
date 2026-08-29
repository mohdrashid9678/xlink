package analytics

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	ClicksStreamKey = "xlink:events:clicks"
	MaxStreamLen    = 100000
)

type StreamProducer interface {
	PublishClick(ctx context.Context, event models.ClickEvent) error
}

type RedisStreamProducer struct {
	client *redis.Client
}

func NewRedisStreamProducer(client *redis.Client) *RedisStreamProducer {
	return &RedisStreamProducer{client: client}
}

func (p *RedisStreamProducer) PublishClick(ctx context.Context, event models.ClickEvent) error {
	if p.client == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal click event: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: ClicksStreamKey,
		MaxLen: MaxStreamLen,
		Approx: true,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	}

	return p.client.XAdd(ctx, args).Err()
}

type NoopProducer struct{}

func NewNoopProducer() *NoopProducer {
	return &NoopProducer{}
}

func (n *NoopProducer) PublishClick(_ context.Context, _ models.ClickEvent) error {
	return nil
}
