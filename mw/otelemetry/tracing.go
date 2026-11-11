// Package otelemetry provides OpenTelemetry-based distributed tracing middleware for
// ziggurat, following the same shape as mw/prometheus and mw/event: a Middleware
// (ziggurat.Use-compatible) that wraps a ziggurat.Handler.
package otelemetry

import (
	"fmt"
	"strconv"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/internal/eventmeta"

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

// HeaderReplay marks a message that was moved back from a dead-letter queue by a
// replay. TraceHandler starts a NEW trace for such a message and only links it to the
// original trace, because a replay can happen days after the original consume and a
// parent/child relationship would stretch the original trace across that gap.
const HeaderReplay = "ziggurat-replay"

// instrumentationName is the tracer used for spans created outside TraceHandler
// (see StartPublishSpan).
const instrumentationName = "github.com/gojekfarm/ziggurat/v2/mw/otelemetry"

// retryCountMetadataKey mirrors mw/rabbitmq.KeyRetryCount, which this package cannot
// import (mw/rabbitmq imports this package).
const retryCountMetadataKey = "rabbitmqAutoRetryCount"

// consumeSpanInfo derives the span name and attributes for a consumed event.
//
// The span NAME is low-cardinality - "<topic> process" for Kafka, "<queue> process"
// for RabbitMQ - so spans group sensibly in a tracing backend. Everything that varies
// per partition/message goes in attributes. When neither topic nor queue is known the
// raw routing path is used, which is what earlier versions always did.
func consumeSpanInfo(event *ziggurat.Event) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{
		attribute.String("messaging.operation", "process"),
		// Deprecated in favour of messaging.destination.name; kept so existing trace
		// queries keep working for one more release.
		attribute.String("messaging.destination", event.RoutingPath),
		attribute.String("ziggurat.routing_path", event.RoutingPath),
	}

	if queue, ok := eventmeta.RabbitMQQueue(event); ok {
		attrs = append(attrs,
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", queue),
		)
		// A retried event is the original event, re-serialised: EventType still says
		// where it first came from.
		if event.EventType != "" {
			attrs = append(attrs, attribute.String("ziggurat.event_origin", event.EventType))
		}
		if n, ok := retryCount(event); ok {
			attrs = append(attrs, attribute.Int("ziggurat.retry.count", n))
		}
		return queue + " process", attrs
	}

	attrs = append(attrs, attribute.String("messaging.system", event.EventType))
	if group, topic, partition, ok := eventmeta.Kafka(event); ok {
		attrs = append(attrs,
			attribute.String("messaging.destination.name", topic),
			attribute.String("messaging.kafka.consumer.group", group),
			attribute.String("messaging.kafka.destination.partition", strconv.Itoa(partition)),
		)
		return topic + " process", attrs
	}
	return event.RoutingPath, attrs
}

// retryCount reads the retry count mw/rabbitmq stores in Event.Metadata. After a JSON
// round trip through RabbitMQ it is a float64, before that an int.
func retryCount(event *ziggurat.Event) (int, bool) {
	switch v := event.Metadata[retryCountMetadataKey].(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

// TraceHandler returns Middleware that starts a CONSUMER span for every message
// handled, as a child of any parent trace context found in the message's headers - so a
// message published as part of an already-traced operation (e.g. a retry/DLQ republish,
// see mw/rabbitmq/publish.go) continues the same trace instead of starting a new,
// disconnected one. See consumeSpanInfo for span naming and attributes.
//
// Exception: a message carrying the HeaderReplay header (a DLQ replay) starts a new root
// trace with a span link to the original, so replays done long after the fact do not
// stretch the original trace.
//
// A panic inside next.Handle is recorded on the span (as an error, with span status set
// to Error and error.type=panic) and then re-panicked - this middleware observes, it
// never swallows. The framework's core consume paths deliberately do not recover handler
// panics (see CHANGELOG); this middleware must not reintroduce that behavior as a side
// effect of adding tracing.
//
// Sampling is controlled entirely by whatever TracerProvider is registered via
// otel.SetTracerProvider (see Setup in this package) - not by this middleware.
func TraceHandler(tracerName string) func(ziggurat.Handler) ziggurat.Handler {
	tracer := otel.Tracer(tracerName)

	return func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
			carrier := headersFrom(event)
			ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

			name, attrs := consumeSpanInfo(event)
			startOpts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attrs...),
			}
			if carrier[HeaderReplay] == "true" {
				if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
					startOpts = append(startOpts,
						trace.WithNewRoot(),
						trace.WithLinks(trace.Link{SpanContext: sc}),
						trace.WithAttributes(attribute.Bool("ziggurat.replay", true)),
					)
				}
			}

			ctx, span := tracer.Start(ctx, name, startOpts...)
			defer span.End()

			defer func() {
				if r := recover(); r != nil {
					span.RecordError(fmt.Errorf("panic: %v", r))
					span.SetAttributes(attribute.String("error.type", "panic"))
					span.SetStatus(codes.Error, "panic")
					panic(r) // re-panic: this middleware observes, never swallows
				}
			}()

			next.Handle(ctx, event)
		})
	}
}

// StartPublishSpan starts a PRODUCER span named "<destination> publish" for a message
// being published to a RabbitMQ queue, with the given extra attributes. Inject the
// returned context (not the caller's) into the outgoing headers, so the consumer of that
// message becomes a child of this publish span - which is what makes retry lineage
// visible as one trace: consume -> publish(retry) -> consume -> publish(dlq).
//
// The caller must call span.End(), and should RecordError/SetStatus on failure.
func StartPublishSpan(ctx context.Context, destination string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	base := []attribute.KeyValue{
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination.name", destination),
		attribute.String("messaging.operation", "publish"),
	}
	return otel.Tracer(instrumentationName).Start(ctx, destination+" publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(append(base, attrs...)...),
	)
}

// InjectTraceContext injects the span context active on ctx into carrier (e.g. a
// HeaderCarrier wrapping an outgoing message's headers map), so a republished message
// (retry/DLQ) carries trace lineage back to the span that's active when it's published.
// See mw/rabbitmq/publish.go for where this is called on the retry/DLQ publish path.
func InjectTraceContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}
