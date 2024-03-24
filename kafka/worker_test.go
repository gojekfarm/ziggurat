package kafka

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/stretchr/testify/mock"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {

	t.Run("worker processes messages successfully", func(t *testing.T) {
		var msgCount int32
		mc := MockConsumer{}
		w := worker{
			handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
				atomic.AddInt32(&msgCount, 1)
				return nil
			}),
			logger:      logger.NOOP,
			consumer:    &mc,
			routeGroup:  "foo",
			pollTimeout: 100,
			killSig:     make(chan struct{}),
			id:          "foo-worker",
		}

		topic := "foo"
		mc.On("Logs").Return(make(chan kafka.LogEvent))
		mc.On("Poll", 100).Return(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 1},
		})
		mc.On("StoreOffsets", mock.AnythingOfType("[]kafka.TopicPartition")).Return([]kafka.TopicPartition{}, nil)
		mc.On("Close").Return(nil)
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		w.run(ctx)
		if !errors.Is(w.err, context.DeadlineExceeded) {
			t.Errorf("expected error to be nil got:%v", w.err)
			return
		}
		if msgCount < 1 {
			t.Error("handler was never invoked")
		}

	})

	t.Run("confluent consumer returns a fatal error", func(t *testing.T) {
		mc := MockConsumer{}
		w := worker{
			handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
				return nil
			}),
			logger:      logger.NOOP,
			consumer:    &mc,
			routeGroup:  "foo",
			pollTimeout: 100,
			killSig:     make(chan struct{}),
			id:          "foo-worker",
		}

		mc.On("Logs").Return(make(chan kafka.LogEvent))
		mc.On("Poll", 100).Return(kafka.NewError(kafka.ErrAllBrokersDown, "fatal error", true))
		mc.On("Close").Return(nil)
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		w.run(ctx)

		if errors.Is(w.err, context.DeadlineExceeded) {
			t.Error("expected a kafka error")
			return
		}
		t.Logf("error:%v", w.err)
	})

	t.Run("worker kill", func(t *testing.T) {
		mc := MockConsumer{}
		w := worker{
			handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
				return nil
			}),
			logger:      logger.NOOP,
			consumer:    &mc,
			routeGroup:  "foo",
			pollTimeout: 100,
			killSig:     make(chan struct{}),
			id:          "foo-worker",
		}

		topic := "foo"
		mc.On("Logs").Return(make(chan kafka.LogEvent))
		mc.On("StoreOffsets", mock.AnythingOfType("[]kafka.TopicPartition")).Return([]kafka.TopicPartition{}, nil)
		mc.On("Poll", 100).Return(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 1},
		})
		mc.On("Close").Return(nil)

		timer := time.AfterFunc(500*time.Millisecond, func() {
			w.kill()
		})
		defer timer.Stop()

		w.run(context.Background())

		var expectedErr ErrorWorkerKilled
		if found := errors.As(w.err, &expectedErr); !found {
			t.Error("expected worker killed error got:", w.err.Error())
			return
		}

	})

}
