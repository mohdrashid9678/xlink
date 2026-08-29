package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	ConsumerGroupName = "xlink-analytics-group"
	DefaultBatchSize  = 500
	DefaultPollWait   = 1 * time.Second
)

type ConsumerConfig struct {
	BatchSize    int64
	PollTimeout  time.Duration
	ConsumerName string
}

type StreamConsumer struct {
	client     *redis.Client
	repository repository.AnalyticsRepository
	logger     *slog.Logger
	cfg        ConsumerConfig
	wg         sync.WaitGroup
	cancel     context.CancelFunc
}

func NewStreamConsumer(
	client *redis.Client,
	repository repository.AnalyticsRepository,
	logger *slog.Logger,
	cfg ConsumerConfig,
) *StreamConsumer {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultBatchSize
	}
	if cfg.PollTimeout <= 0 {
		cfg.PollTimeout = DefaultPollWait
	}
	if cfg.ConsumerName == "" {
		cfg.ConsumerName = fmt.Sprintf("consumer-%s", uuid.New().String()[:8])
	}

	return &StreamConsumer{
		client:     client,
		repository: repository,
		logger:     logger,
		cfg:        cfg,
	}
}

func (c *StreamConsumer) Start(ctx context.Context) error {
	if c.client == nil || c.repository == nil {
		return nil
	}

	// 1. Ensure Consumer Group exists
	err := c.client.XGroupCreateMkStream(ctx, ClicksStreamKey, ConsumerGroupName, "0").Err()
	if err != nil && !isBusyGroupError(err) {
		return fmt.Errorf("create redis stream consumer group: %w", err)
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.wg.Add(1)
	go c.workerLoop(consumerCtx)

	return nil
}

func (c *StreamConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *StreamConsumer) workerLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.processBatch(ctx)
		}
	}
}

func (c *StreamConsumer) processBatch(ctx context.Context) {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroupName,
		Consumer: c.cfg.ConsumerName,
		Streams:  []string{ClicksStreamKey, ">"},
		Count:    c.cfg.BatchSize,
		Block:    c.cfg.PollTimeout,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
			return
		}
		if c.logger != nil {
			c.logger.Error("Error reading clickstream from Redis", "error", err)
		}
		time.Sleep(500 * time.Millisecond)
		return
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return
	}

	messages := streams[0].Messages
	events := make([]models.ClickEvent, 0, len(messages))
	ackIDs := make([]string, 0, len(messages))

	for _, msg := range messages {
		payloadStr, ok := msg.Values["payload"].(string)
		if !ok {
			ackIDs = append(ackIDs, msg.ID)
			continue
		}

		var event models.ClickEvent
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			if c.logger != nil {
				c.logger.Warn("Failed to unmarshal click event payload", "msg_id", msg.ID, "error", err)
			}
			ackIDs = append(ackIDs, msg.ID)
			continue
		}

		events = append(events, event)
		ackIDs = append(ackIDs, msg.ID)
	}

	if len(events) > 0 {
		if err := c.repository.BatchInsertClicks(ctx, events); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to persist batch click events to database", "count", len(events), "error", err)
			}
			// Do not ACK on DB failure so Redis Streams can retry or replay
			return
		}
	}

	if len(ackIDs) > 0 {
		_ = c.client.XAck(ctx, ClicksStreamKey, ConsumerGroupName, ackIDs...).Err()
	}
}

func isBusyGroupError(err error) bool {
	return err != nil && (err.Error() == "BUSYGROUP Consumer Group name already exists" ||
		err.Error() == "BUSYGROUP")
}
