package kafka

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockHandler struct {
	mock.Mock
}

func (mh *mockHandler) Handle(ctx context.Context, event *ziggurat.Event) {
	return
}

func TestWorker(t *testing.T) {

	t.Run("worker processes messages successfully", func(t *testing.T) {

		wantEvent := ziggurat.Event{
			RoutingPath: "foo-group/foo/1",
			EventType:   "kafka",
			Value:       make([]byte, 0),
			Key:         make([]byte, 0),
			Metadata:    map[string]any{"kafka-partition": 1, "kafka-topic": "foo"},
		}

		eventMatcher := mock.MatchedBy(func(e *ziggurat.Event) bool {
			diff := cmp.Diff(&wantEvent, e, cmpopts.IgnoreFields(ziggurat.Event{}, "ReceivedTimestamp"))
			if diff != "" {
				t.Logf("(-Want +Got)%s\n", diff)
				return false
			}
			return true

		})
		mh := mockHandler{}
		mh.On("Handle", mock.Anything, eventMatcher).Return(nil)

		mc := MockConsumer{}
		w := worker{
			handler:     &mh,
			logger:      logger.NOOP,
			consumer:    &mc,
			routeGroup:  "foo-group",
			pollTimeout: 100,
			killSig:     make(chan struct{}),
			id:          "foo-worker",
		}

		topic := "foo"
		mc.On("Logs").Return(make(chan kafka.LogEvent))
		mc.On("Poll", 100).Return(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 1},
		})
		mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
		mc.On("StoreOffsets", mock.AnythingOfType("[]kafka.TopicPartition")).Return([]kafka.TopicPartition{}, nil)
		mc.On("Close").Return(nil)
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		w.run(ctx)
		if !errors.Is(w.err, context.DeadlineExceeded) {
			t.Errorf("expected error to be nil got:%v", w.err)
			return
		}

	})

	t.Run("confluent consumer returns a fatal error", func(t *testing.T) {
		mc := MockConsumer{}
		w := worker{
			handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {

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
		mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
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
			handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
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
		mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
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
