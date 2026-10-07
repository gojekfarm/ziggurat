// Package otelemetry provides OpenTelemetry-based distributed tracing middleware for
// ziggurat, following the same shape as mw/prometheus and mw/event: a Middleware
// (ziggurat.Use-compatible) that wraps a ziggurat.Handler.
package otelemetry

import (
	"fmt"

	"github.com/gojekfarm/ziggurat/v2"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"context"
)

// HeaderCarrier adapts the map[string]string stored at
// Event.Metadata[ziggurat.MetadataKeyHeaders] (populated by kafka/process.go and
// mw/rabbitmq/consumer.go from the native transport headers) to OTel's
// propagation.TextMapCarrier, so a W3C traceparent/tracestate can be extracted from -
// and, on republish, injected into - a message's headers.
type HeaderCarrier map[string]string

func (c HeaderCarrier) Get(key string) string { return c[key] }

func (c HeaderCarrier) Set(key, value string) { c[key] = value }

func (c HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// headersFrom returns the message headers already captured on the event as a
// HeaderCarrier, or a fresh empty one if none were captured (e.g. the transport didn't
// set any, or this event didn't originate from kafka/rabbitmq's header-capturing code).
func headersFrom(event *ziggurat.Event) HeaderCarrier {
	if event.Metadata != nil {
		if h, ok := event.Metadata[ziggurat.MetadataKeyHeaders].(map[string]string); ok && h != nil {
			return HeaderCarrier(h)
		}
	}
	return HeaderCarrier{}
}

// TraceHandler returns Middleware that starts a span for every message handled, named
// after the event's RoutingPath, as a child of any parent trace context found in the
// message's headers - so a message published as part of an already-traced operation
// (e.g. a retry/DLQ republish, see mw/rabbitmq/publish.go) continues the same trace
// instead of starting a new, disconnected one.
//
// A panic inside next.Handle is recorded on the span (as an error, with span status set
// to Error) and then re-panicked - this middleware observes, it never swallows. The
// framework's core consume paths deliberately do not recover handler panics (see
// CHANGELOG); this middleware must not reintroduce that behavior as a side effect of
// adding tracing.
//
// Sampling is controlled entirely by whatever TracerProvider is registered via
// otel.SetTracerProvider (see Setup in this package) - not by this middleware.
func TraceHandler(tracerName string) func(ziggurat.Handler) ziggurat.Handler {
	tracer := otel.Tracer(tracerName)

	return func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
			carrier := headersFrom(event)
			ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

			ctx, span := tracer.Start(ctx, event.RoutingPath, trace.WithAttributes(
				attribute.String("messaging.system", event.EventType),
				attribute.String("messaging.destination", event.RoutingPath),
			))
			defer span.End()

			defer func() {
				if r := recover(); r != nil {
					span.RecordError(fmt.Errorf("panic: %v", r))
					span.SetStatus(codes.Error, "panic")
					panic(r) // re-panic: this middleware observes, never swallows
				}
			}()

			next.Handle(ctx, event)
		})
	}
}

// InjectTraceContext injects the span context active on ctx into carrier (e.g. a
// HeaderCarrier wrapping an outgoing message's headers map), so a republished message
// (retry/DLQ) carries trace lineage back to the span that's active when it's published.
// See mw/rabbitmq/publish.go for where this is called on the retry/DLQ publish path.
func InjectTraceContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}
