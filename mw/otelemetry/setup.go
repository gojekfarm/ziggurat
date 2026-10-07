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
// Sampling is AlwaysSample for this initial rollout - every message gets a span. This
// is the simplest option to start with; if Kafka throughput makes this too expensive,
// switch to sdktrace.ParentBased(sdktrace.TraceIDRatioBased(rate)) here later - nothing
// else in this package needs to change to adopt a different sampler.
//
// The returned shutdown func flushes any buffered spans and closes the exporter
// connection; call it during graceful shutdown (e.g. deferred in main, or alongside
// whatever already handles SIGTERM for the consumer process).
func Setup(ctx context.Context, serviceName, otlpEndpoint string) (shutdown func(context.Context) error, err error) {
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
