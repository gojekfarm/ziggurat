package kafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

// confluentConsumer is the subset of *kafka.Consumer that ziggurat's worker and
// rebalance callback actually need. The interface is intentionally narrow so that
// tests can substitute a MockConsumer without having to stub methods they don't use.
//
// IncrementalAssign / IncrementalUnassign are the cooperative-rebalance APIs:
// instead of replacing the full partition assignment in one shot (eager), they add
// or remove only the deltas the coordinator hands to this consumer instance. They
// must be called from the rebalance callback (see createConsumer) when the consumer
// group uses "cooperative-sticky" (or any other COOPERATIVE_ protocol).
//
// CommitOffsets performs a synchronous offset commit for an explicit list of
// topic-partitions, bypassing the auto-commit timer. The rebalance callback calls
// it on RevokedPartitions so that in-memory stored offsets are flushed before the
// partitions are handed back to the coordinator, preventing duplicate processing
// after a membership change.
type confluentConsumer interface {
	Poll(int) kafka.Event
	StoreOffsets([]kafka.TopicPartition) ([]kafka.TopicPartition, error)
	Logs() chan kafka.LogEvent
	Commit() ([]kafka.TopicPartition, error)
	CommitOffsets([]kafka.TopicPartition) ([]kafka.TopicPartition, error)
	IncrementalAssign([]kafka.TopicPartition) error
	IncrementalUnassign([]kafka.TopicPartition) error
	Close() error
}

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) Poll(i int) kafka.Event {
	args := m.Called(i)
	time.Sleep(time.Duration(i) * time.Millisecond)
	return args.Get(0).(kafka.Event)
}

func (m *MockConsumer) StoreOffsets(partitions []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	args := m.Called(partitions)
	return args.Get(0).([]kafka.TopicPartition), args.Error(1)
}

func (m *MockConsumer) Commit() ([]kafka.TopicPartition, error) {
	args := m.Called()
	return args.Get(0).([]kafka.TopicPartition), args.Error(1)
}

func (m *MockConsumer) CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	args := m.Called(offsets)
	return args.Get(0).([]kafka.TopicPartition), args.Error(1)
}

func (m *MockConsumer) IncrementalAssign(partitions []kafka.TopicPartition) error {
	return m.Called(partitions).Error(0)
}

func (m *MockConsumer) IncrementalUnassign(partitions []kafka.TopicPartition) error {
	return m.Called(partitions).Error(0)
}

func (m *MockConsumer) Close() error {
	return m.Called().Error(0)
}

func (m *MockConsumer) Logs() chan kafka.LogEvent {
	return m.Called().Get(0).(chan kafka.LogEvent)
}
