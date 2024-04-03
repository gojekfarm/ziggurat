package kafka

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"sync/atomic"
	"testing"
	"time"
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
				ConsumerCount:    5,
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

}

func makePtr[V any](v V) *V {
	return &v
}
