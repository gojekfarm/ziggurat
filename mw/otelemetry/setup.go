package otelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup configures the process-wide OpenTelemetry TracerProvider to export spans via
// OTLP/gRPC to the given collector endpoint (e.g. "localhost:4317" for a local OTel
// Collector - see docker-compose-tracing.yml), registers it as the global
// TracerProvider (what otel.Tracer(...) and TraceHandler use), and registers the W3C
// TraceContext + Baggage propagators (what TraceHandler/InjectTraceContext use to
// extract/inject via message headers).
//
// Tracing is opt-in: enabled controls whether any of this actually happens. Callers are
// expected to source this from their own config/env var (e.g.
// `enabled := os.Getenv("TRACING_ENABLED") == "true"`), which gives tracing a sensible
// off-by-default posture for free, since an unset env var yields false. When enabled is
// false, Setup does no OTLP dial, registers nothing, and returns a no-op shutdown and a
// nil error - safe to call unconditionally from main regardless of the flag's value.
// TraceHandler still works when tracing is disabled: it runs against whatever
// TracerProvider/propagator are currently registered globally, which is otel's own
// no-op implementation unless something has called otel.SetTracerProvider - so spans
// are created but discarded, and header extraction/injection is a cheap no-op, with no
// network calls.
//
// Sampling is AlwaysSample for this initial rollout - every message gets a span. This
// is the simplest option to start with; if Kafka throughput makes this too expensive,
// switch to sdktrace.ParentBased(sdktrace.TraceIDRatioBased(rate)) here later - nothing
// else in this package needs to change to adopt a different sampler.
//
// The returned shutdown func flushes any buffered spans and closes the exporter
// connection; call it during graceful shutdown (e.g. deferred in main, or alongside
// whatever already handles SIGTERM for the consumer process). It is always safe to call,
// even when enabled is false.
func Setup(ctx context.Context, serviceName, otlpEndpoint string, enabled bool) (shutdown func(context.Context) error, err error) {
	noopShutdown := func(context.Context) error { return nil }

	if !enabled {
		return noopShutdown, nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(), // local collector, no TLS - revisit if the collector endpoint moves off-box
	)
	if err != nil {
		return nil, fmt.Errorf("otelemetry: failed to create OTLP exporter: %w", err)
	}

	// NewSchemaless + a plain attribute (rather than the semconv package's
	// ServiceNameKey) deliberately avoids pinning this package to one specific
	// semconv schema version - "service.name" is a stable, won't-change attribute key.
	res := resource.NewSchemaless(attribute.String("service.name", serviceName))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
