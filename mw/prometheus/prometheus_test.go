package prometheus

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPublishHandlerMetrics_Success(t *testing.T) {
	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {})
	wrapped := PublishHandlerMetrics(handler)

	event := &ziggurat.Event{RoutingPath: "test/success/path"}
	wrapped.Handle(context.Background(), event)

	got := testutil.ToFloat64(HandlerEventsCounter.With(map[string]string{
		RouteLabel:   event.RoutingPath,
		OutcomeLabel: OutcomeSuccess,
	}))
	if got != 1 {
		t.Errorf("expected events counter with outcome=success to be 1, got %v", got)
	}
}

// TestPublishHandlerMetrics_PanicIsObservedNotSwallowed is the regression test for the
// core requirement of this middleware: it must observe a panic (record outcome=panic)
// but never swallow it. The framework's consume paths deliberately do not recover
// handler panics anymore (see CHANGELOG) - this middleware re-introducing swallowing
// behavior as a side effect of adding metrics would be a regression of that decision.
func TestPublishHandlerMetrics_PanicIsObservedNotSwallowed(t *testing.T) {
	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		panic("deliberate test panic")
	})
	wrapped := PublishHandlerMetrics(handler)

	event := &ziggurat.Event{RoutingPath: "test/panic/path"}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected the panic to propagate out of PublishHandlerMetrics, but it was swallowed")
		}
		if r != "deliberate test panic" {
			t.Errorf("expected the original panic value to be preserved through re-panic, got: %v", r)
		}

		got := testutil.ToFloat64(HandlerEventsCounter.With(map[string]string{
			RouteLabel:   event.RoutingPath,
			OutcomeLabel: OutcomePanic,
		}))
		if got != 1 {
			t.Errorf("expected events counter with outcome=panic to be 1, got %v", got)
		}
	}()

	wrapped.Handle(context.Background(), event)
}
