package kafka

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/logger"
)

func TestWorkerOrchestration(t *testing.T) {
	t.Run("workers do not error out", func(t *testing.T) {
		mc := MockConsumer{}
		cg := ConsumerGroup{
			Logger: logger.NewLogger(logger.LevelError),
			GroupConfig: ConsumerConfig{
				BootstrapServers: "localhost:9092",
				GroupID:          "group-test",
				Topics:           []string{"foo"},
				ConsumerCount:    1,
			},
			consumerMakeFunc: func(configMap *kafka.ConfigMap, strings []string) confluentConsumer {
				return &mc
			},
		}

		logChan := make(chan kafka.LogEvent)
		expectedTopicPart := kafka.TopicPartition{Topic: makePtr("foo"), Partition: 1}
		expectedTopicPartStoreOffsets := kafka.TopicPartition{Topic: makePtr("foo"), Partition: 1, Offset: 1}
		mc.On("Poll", 100).Return(&kafka.Message{
			TopicPartition: expectedTopicPart,
		})
		mc.On("StoreOffsets", []kafka.TopicPartition{expectedTopicPartStoreOffsets}).
			Return([]kafka.TopicPartition{expectedTopicPartStoreOffsets}, nil)
		mc.On("Close").Return(nil)
		mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
		mc.On("Logs").Return(logChan)

		var msgCount int32
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err := cg.Consume(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
			atomic.AddInt32(&msgCount, 1)
		}))
		if !errors.Is(err, ErrCleanShutdown) {
			t.Error("expected nil error, got:", err.Error())
			return
		}
		if atomic.LoadInt32(&msgCount) < 1 {
			t.Error("expected a non zero message count")
		}
	})

	t.Run("consumerMakeFunc is called once per worker", func(t *testing.T) {
		const consumerCount = 5
		var makeCount int32

		cg := ConsumerGroup{
			Logger: logger.NewLogger(logger.LevelError),
			GroupConfig: ConsumerConfig{
				BootstrapServers: "localhost:9092",
				GroupID:          "group-test",
				Topics:           []string{"foo"},
				ConsumerCount:    consumerCount,
			},
			consumerMakeFunc: func(configMap *kafka.ConfigMap, topics []string) confluentConsumer {
				atomic.AddInt32(&makeCount, 1)
				return stubConsumerForOrchestration()
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		err := cg.Consume(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {}))
		if !errors.Is(err, ErrCleanShutdown) {
			t.Fatalf("expected clean shutdown, got: %v", err)
		}
		if atomic.LoadInt32(&makeCount) != consumerCount {
			t.Fatalf("expected consumerMakeFunc to be called %d times, got %d", consumerCount, makeCount)
		}
	})

	t.Run("each worker gets its own consumer instance", func(t *testing.T) {
		const consumerCount = 3
		mocks := make([]*MockConsumer, 0, consumerCount)

		cg := ConsumerGroup{
			Logger: logger.NewLogger(logger.LevelError),
			GroupConfig: ConsumerConfig{
				BootstrapServers: "localhost:9092",
				GroupID:          "group-test",
				Topics:           []string{"foo"},
				ConsumerCount:    consumerCount,
			},
			consumerMakeFunc: func(configMap *kafka.ConfigMap, topics []string) confluentConsumer {
				mc := stubConsumerForOrchestration()
				mocks = append(mocks, mc)
				return mc
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		err := cg.Consume(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {}))
		if !errors.Is(err, ErrCleanShutdown) {
			t.Fatalf("expected clean shutdown, got: %v", err)
		}
		if len(mocks) != consumerCount {
			t.Fatalf("expected %d consumer instances, got %d", consumerCount, len(mocks))
		}
		seen := make(map[*MockConsumer]struct{}, consumerCount)
		for _, mc := range mocks {
			if _, dup := seen[mc]; dup {
				t.Fatal("expected distinct consumer instance per worker")
			}
			seen[mc] = struct{}{}
			mc.AssertNumberOfCalls(t, "Close", 1)
			mc.AssertNumberOfCalls(t, "Commit", 1)
		}
	})
}

func stubConsumerForOrchestration() *MockConsumer {
	mc := &MockConsumer{}
	logChan := make(chan kafka.LogEvent)
	mc.On("Logs").Return(logChan)
	mc.On("Poll", 100).Maybe().Return(kafka.Event(nil))
	mc.On("Close").Return(nil)
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	return mc
}

func makePtr[V any](v V) *V {
	return &v
}
