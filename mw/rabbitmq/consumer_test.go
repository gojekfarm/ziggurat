package rabbitmq

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/streadway/amqp"
)

// fakeAcknowledger is a minimal amqp.Acknowledger so amqp.Delivery{Ack,Reject,Nack}
// can be exercised directly in a test, without needing a live RabbitMQ connection.
type fakeAcknowledger struct {
	mu       sync.Mutex
	acked    bool
	rejected bool
	requeue  bool
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.acked = true
	return nil
}

func (f *fakeAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejected = true
	f.requeue = requeue
	return nil
}

func (f *fakeAcknowledger) wasAcked() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.acked
}

func (f *fakeAcknowledger) wasRejected() (rejected, requeue bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rejected, f.requeue
}

// fakeStructuredLogger is a minimal ziggurat.StructuredLogger that records Error
// calls. Named distinctly from pubpool_test.go's "fakeLogger" (which implements the
// unrelated amqpextra logger.Logger/Printf interface) to avoid a name collision.
type fakeStructuredLogger struct {
	mu       sync.Mutex
	errCalls []struct {
		message string
		err     error
	}
}

func (f *fakeStructuredLogger) Info(string, ...map[string]interface{})        {}
func (f *fakeStructuredLogger) Debug(string, ...map[string]interface{})       {}
func (f *fakeStructuredLogger) Warn(string, ...map[string]interface{})        {}
func (f *fakeStructuredLogger) Fatal(string, error, ...map[string]interface{}) {}

func (f *fakeStructuredLogger) Error(message string, err error, _ ...map[string]interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errCalls = append(f.errCalls, struct {
		message string
		err     error
	}{message, err})
}

func (f *fakeStructuredLogger) errorCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.errCalls)
}

func (f *fakeStructuredLogger) lastError() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.errCalls) == 0 {
		return "", nil
	}
	last := f.errCalls[len(f.errCalls)-1]
	return last.message, last.err
}

func newTestDelivery(t *testing.T, event ziggurat.Event, ack *fakeAcknowledger) amqp.Delivery {
	t.Helper()
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal test event: %v", err)
	}
	return amqp.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		Body:         body,
	}
}

func TestHandleDelivery_RecoversFromHandlerPanic(t *testing.T) {
	ack := &fakeAcknowledger{}
	fl := &fakeStructuredLogger{}
	msg := newTestDelivery(t, ziggurat.Event{Key: []byte("TRIGGER_PANIC")}, ack)

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		if string(e.Key) == "TRIGGER_PANIC" {
			panic("deliberate poison message panic")
		}
	})

	// If handleDelivery does not recover the panic, this call itself panics and fails
	// the test process — the test passing at all is part of the assertion.
	handleDelivery(context.Background(), msg, handler, fl, "test-consumer")

	if !ack.wasAcked() {
		t.Error("expected message to be Acked after a recovered panic (matching happy-path behavior)")
	}
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

func TestHandleDelivery_HandlerDoesNotPanic(t *testing.T) {
	ack := &fakeAcknowledger{}
	fl := &fakeStructuredLogger{}
	var gotEvent *ziggurat.Event

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		gotEvent = e
	})

	msg := newTestDelivery(t, ziggurat.Event{Key: []byte("normal-key")}, ack)
	handleDelivery(context.Background(), msg, handler, fl, "test-consumer")

	if gotEvent == nil || string(gotEvent.Key) != "normal-key" {
		t.Fatalf("expected handler to receive the deserialized event, got: %+v", gotEvent)
	}
	if !ack.wasAcked() {
		t.Error("expected message to be Acked on the happy path")
	}
	if fl.errorCallCount() != 0 {
		t.Errorf("expected no error logs on the happy path, got %d", fl.errorCallCount())
	}
}

func TestHandleDelivery_MalformedMessageIsRejectedWithRequeue(t *testing.T) {
	ack := &fakeAcknowledger{}
	fl := &fakeStructuredLogger{}

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		t.Error("handler should not be invoked for a message that fails to deserialize")
	})

	msg := amqp.Delivery{Acknowledger: ack, DeliveryTag: 1, Body: []byte("not valid json")}
	handleDelivery(context.Background(), msg, handler, fl, "test-consumer")

	rejected, requeue := ack.wasRejected()
	if !rejected || !requeue {
		t.Errorf("expected a malformed message to be rejected with requeue=true (existing behavior, unchanged by this fix), got rejected=%v requeue=%v", rejected, requeue)
	}
	if ack.wasAcked() {
		t.Error("a malformed message should not be Acked")
	}
}
