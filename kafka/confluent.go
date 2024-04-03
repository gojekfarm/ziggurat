package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
	"time"
)

type confluentConsumer interface {
	Poll(int) kafka.Event
	StoreOffsets([]kafka.TopicPartition) ([]kafka.TopicPartition, error)
	Logs() chan kafka.LogEvent
	Commit() ([]kafka.TopicPartition, error)
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

func (m *MockConsumer) Close() error {
	return m.Called().Error(0)
}

func (m *MockConsumer) Logs() chan kafka.LogEvent {
	return m.Called().Get(0).(chan kafka.LogEvent)
}
