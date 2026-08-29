package analytics

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnalyticsRepoForConsumer struct {
	mock.Mock
}

func (m *mockAnalyticsRepoForConsumer) BatchInsertClicks(ctx context.Context, clicks []models.ClickEvent) error {
	args := m.Called(ctx, clicks)
	return args.Error(0)
}

func (m *mockAnalyticsRepoForConsumer) GetAnalyticsSummary(ctx context.Context, urlID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error) {
	panic("unexpected")
}

func TestStreamConsumerWithNilClient(t *testing.T) {
	repo := new(mockAnalyticsRepoForConsumer)
	consumer := NewStreamConsumer(nil, repo, nil, ConsumerConfig{})

	err := consumer.Start(context.Background())
	assert.NoError(t, err)

	consumer.Stop()
}

func TestConsumerConfigDefaults(t *testing.T) {
	consumer := NewStreamConsumer(nil, nil, nil, ConsumerConfig{})
	assert.Equal(t, int64(DefaultBatchSize), consumer.cfg.BatchSize)
	assert.Equal(t, DefaultPollWait, consumer.cfg.PollTimeout)
	assert.NotEmpty(t, consumer.cfg.ConsumerName)
}

func TestIsBusyGroupError(t *testing.T) {
	assert.True(t, isBusyGroupError(errors.New("BUSYGROUP Consumer Group name already exists")))
	assert.True(t, isBusyGroupError(errors.New("BUSYGROUP")))
	assert.False(t, isBusyGroupError(errors.New("ERR syntax error")))
	assert.False(t, isBusyGroupError(nil))
}
