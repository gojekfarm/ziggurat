package kafka

import (
	"context"
	"errors"
	"strings"
	"sync"
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

func TestConsumerGroup_FatalErrorIsNotSwallowed(t *testing.T) {
	mc := MockConsumer{}
	cg := ConsumerGroup{
		Logger: logger.NewLogger(logger.LevelError),
		GroupConfig: ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "group-test-fatal",
			Topics:           []string{"foo"},
			ConsumerCount:    1,
		},
		consumerMakeFunc: func(configMap *kafka.ConfigMap, strings []string) confluentConsumer {
			return &mc
		},
	}

	logChan := make(chan kafka.LogEvent)
	fatalErr := kafka.NewError(kafka.ErrAllBrokersDown, "all brokers down", true) // fatal = true

	mc.On("Poll", 100).Return(fatalErr)
	mc.On("Close").Return(nil)
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("Logs").Return(logChan)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := cg.Consume(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {}))

	// 1. Must not be nil, and must not be the "everything's fine" sentinel — a fatal
	// broker error is neither of those things.
	if err == nil {
		t.Fatal("expected a non-nil error for a fatal broker failure, got nil")
	}
	if errors.Is(err, ErrCleanShutdown) {
		t.Fatalf("Consume reported ErrCleanShutdown despite a fatal broker error — the error was silently swallowed. Got: %v", err)
	}

	// 2. The returned error must actually wrap the underlying kafka.Error — not just be
	// some unrelated non-nil error — so callers using errors.As can still recover the
	// original fatal error and inspect it (e.g. its code, IsFatal()).
	var kErr kafka.Error
	if !errors.As(err, &kErr) {
		t.Fatalf("expected returned error to wrap a kafka.Error via errors.As, got: %v (%T)", err, err)
	}
	if !kErr.IsFatal() {
		t.Errorf("expected the wrapped kafka.Error to still report IsFatal() == true, got false: %v", kErr)
	}
	if kErr.Code() != kafka.ErrAllBrokersDown {
		t.Errorf("expected wrapped error code %v, got %v", kafka.ErrAllBrokersDown, kErr.Code())
	}

	// 3. The error message should still be attributable to the specific worker that failed.
	if !strings.Contains(err.Error(), "group-test-fatal_0") {
		t.Errorf("expected error message to mention the failing worker id, got: %v", err)
	}
}

// TestConsumerGroup_EachWorkerGetsOwnConsumerInstance is the regression test for the
// shared-consumer concurrency bug: previously, ONE confluentConsumer was created outside
// the worker-spawn loop and handed to every worker goroutine, so ConsumerCount workers
// were all calling Poll/Commit/StoreOffsets/Close concurrently on a single, non-thread-safe
// client handle. The fix moves consumerMakeFunc's call inside the loop so each worker gets
// its own, independent consumer. This test asserts both halves of that: consumerMakeFunc is
// invoked once per worker (not once total), and no two workers end up sharing a pointer.
func TestConsumerGroup_EachWorkerGetsOwnConsumerInstance(t *testing.T) {
	var mu sync.Mutex
	var created []*MockConsumer

	const consumerCount = 3

	cg := ConsumerGroup{
		Logger: logger.NewLogger(logger.LevelError),
		GroupConfig: ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "group-independent-consumers",
			Topics:           []string{"foo"},
			ConsumerCount:    consumerCount,
		},
		consumerMakeFunc: func(configMap *kafka.ConfigMap, strings []string) confluentConsumer {
			mu.Lock()
			defer mu.Unlock()

			mc := &MockConsumer{}
			logChan := make(chan kafka.LogEvent)
			// No real messages: Poll returns a non-fatal kafka.Error, which worker.run
			// just logs and keeps polling on (see the `case kafka.Error` branch in
			// worker.go) — harmless busy-polling until ctx is cancelled. A literal nil
			// can't be used here: MockConsumer.Poll does a single-return type assertion
			// (`args.Get(0).(kafka.Event)`), which panics on a nil interface value.
			mc.On("Poll", 100).Return(kafka.NewError(kafka.ErrNoError, "idle poll", false))
			mc.On("Close").Return(nil)
			mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
			mc.On("Logs").Return(logChan)

			created = append(created, mc)
			return mc
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = cg.Consume(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {}))

	mu.Lock()
	defer mu.Unlock()

	if len(created) != consumerCount {
		t.Fatalf("expected consumerMakeFunc to be called once per worker (%d times), got %d", consumerCount, len(created))
	}

	seen := make(map[*MockConsumer]bool, len(created))
	for _, mc := range created {
		if seen[mc] {
			t.Fatalf("expected each worker to receive a distinct consumer instance, found a duplicate pointer")
		}
		seen[mc] = true
	}

	if len(cg.workers) != consumerCount {
		t.Fatalf("expected %d workers, got %d", consumerCount, len(cg.workers))
	}
	workerConsumers := make(map[confluentConsumer]bool, len(cg.workers))
	for _, w := range cg.workers {
		if workerConsumers[w.consumer] {
			t.Error("two workers are sharing the same underlying consumer instance — the shared-consumer bug has regressed")
		}
		workerConsumers[w.consumer] = true
	}
}

func makePtr[V any](v V) *V {
	return &v
}
