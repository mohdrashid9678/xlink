package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNoopProducer(t *testing.T) {
	producer := NewNoopProducer()
	err := producer.PublishClick(context.Background(), models.ClickEvent{
		ID:        uuid.New(),
		ShortCode: "test",
		ClickedAt: time.Now(),
	})
	assert.NoError(t, err)
}

func TestRedisStreamProducerWithNilClient(t *testing.T) {
	producer := NewRedisStreamProducer(nil)
	err := producer.PublishClick(context.Background(), models.ClickEvent{
		ID:        uuid.New(),
		ShortCode: "test",
		ClickedAt: time.Now(),
	})
	assert.NoError(t, err)
}
