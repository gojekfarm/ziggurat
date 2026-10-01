package kafka

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
)

// fakeLogger is a minimal ziggurat.StructuredLogger that records Error calls so tests
// can assert on what was logged, without depending on a real logger implementation.
type fakeLogger struct {
	mu       sync.Mutex
	errCalls []struct {
		message string
		err     error
	}
}

func (f *fakeLogger) Info(string, ...map[string]interface{})  {}
func (f *fakeLogger) Debug(string, ...map[string]interface{}) {}
func (f *fakeLogger) Warn(string, ...map[string]interface{})  {}
func (f *fakeLogger) Fatal(string, error, ...map[string]interface{}) {}

func (f *fakeLogger) Error(message string, err error, _ ...map[string]interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errCalls = append(f.errCalls, struct {
		message string
		err     error
	}{message, err})
}

func (f *fakeLogger) errorCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.errCalls)
}

func (f *fakeLogger) lastError() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.errCalls) == 0 {
		return "", nil
	}
	last := f.errCalls[len(f.errCalls)-1]
	return last.message, last.err
}

func newTestMessage(key, value string) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: makePtr("foo"), Partition: 1, Offset: 1},
		Key:            []byte(key),
		Value:          []byte(value),
		Timestamp:      time.Now(),
	}
}

func TestProcessMessage_RecoversFromHandlerPanic(t *testing.T) {
	fl := &fakeLogger{}
	msg := newTestMessage("TRIGGER_PANIC", "payload")

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		if string(e.Key) == "TRIGGER_PANIC" {
			panic("deliberate poison message panic")
		}
	})

	// If processMessage does not recover the panic, this call itself would panic and
	// fail the test process — the test passing at all is part of the assertion.
	processMessage(context.Background(), msg, handler, "test-group", fl)

	if fl.errorCallCount() != 1 {
		t.Fatalf("expected exactly one Error log call after a recovered panic, got %d", fl.errorCallCount())
	}

	message, err := fl.lastError()
	if !strings.Contains(message, "recovered from panic") {
		t.Errorf("expected log message to mention the recovered panic, got: %q", message)
	}
	if err == nil || !strings.Contains(err.Error(), "deliberate poison message panic") {
		t.Errorf("expected logged error to contain the original panic value, got: %v", err)
	}
}

func TestProcessMessage_SubsequentMessagesStillProcessed(t *testing.T) {
	fl := &fakeLogger{}

	var processedKeys []string
	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		if string(e.Key) == "TRIGGER_PANIC" {
			panic("deliberate poison message panic")
		}
		processedKeys = append(processedKeys, string(e.Key))
	})

	// Simulate a worker's poll loop feeding several messages through processMessage,
	// one of which panics — the point of this test is that the panic does not stop
	// the remaining, otherwise-healthy messages from being processed afterward.
	processMessage(context.Background(), newTestMessage("before", "v"), handler, "test-group", fl)
	processMessage(context.Background(), newTestMessage("TRIGGER_PANIC", "v"), handler, "test-group", fl)
	processMessage(context.Background(), newTestMessage("after", "v"), handler, "test-group", fl)

	want := []string{"before", "after"}
	if len(processedKeys) != len(want) || processedKeys[0] != want[0] || processedKeys[1] != want[1] {
		t.Fatalf("expected messages before/after the panic to be processed as %v, got %v", want, processedKeys)
	}
	if fl.errorCallCount() != 1 {
		t.Errorf("expected exactly one recovered-panic log call, got %d", fl.errorCallCount())
	}
}

func TestProcessMessage_HandlerDoesNotPanic(t *testing.T) {
	fl := &fakeLogger{}
	var gotEvent *ziggurat.Event

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		gotEvent = e
	})

	processMessage(context.Background(), newTestMessage("k1", "v1"), handler, "test-group", fl)

	if gotEvent == nil {
		t.Fatal("expected handler to receive an event")
	}
	if string(gotEvent.Key) != "k1" || string(gotEvent.Value) != "v1" {
		t.Errorf("unexpected event key/value: %q / %q", gotEvent.Key, gotEvent.Value)
	}
	if gotEvent.RoutingPath != "test-group/foo/1" {
		t.Errorf("unexpected routing path: %q", gotEvent.RoutingPath)
	}
	if fl.errorCallCount() != 0 {
		t.Errorf("expected no error logs on the happy path, got %d", fl.errorCallCount())
	}
}
