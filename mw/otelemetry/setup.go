package otelemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// DefaultSampleRatio is the fraction of new (root) traces recorded when no sampler is
// configured: ParentBased(TraceIDRatioBased(DefaultSampleRatio)). A remote parent's
// sampling decision is always respected, so a message traced upstream stays traced.
const DefaultSampleRatio = 0.1

// Defaults for the span batcher and exporter. They are explicit (instead of relying on
// the SDK defaults) so a change in SDK defaults cannot silently change memory use or
// export behaviour; override with WithBatchOptions.
const (
	defaultMaxQueueSize       = 4096
	defaultMaxExportBatchSize = 512
	defaultBatchTimeout       = 5 * time.Second
	defaultExportTimeout      = 10 * time.Second
)

type setupConfig struct {
	sampler     sdktrace.Sampler
	tlsConfig   *tls.Config
	version     string
	environment string
	attrs       []attribute.KeyValue
	batchOpts   []sdktrace.BatchSpanProcessorOption
}

// SetupOption configures Setup.
type SetupOption func(*setupConfig)

// WithSampler sets the sampler. Without it (or WithSampleRatio) Setup samples
// DefaultSampleRatio of new traces. Use sdktrace.AlwaysSample() to record everything.
func WithSampler(s sdktrace.Sampler) SetupOption {
	return func(c *setupConfig) { c.sampler = s }
}

// WithSampleRatio records `ratio` (0..1) of new traces and always follows the parent's
// decision when there is one: ParentBased(TraceIDRatioBased(ratio)).
func WithSampleRatio(ratio float64) SetupOption {
	return func(c *setupConfig) {
		c.sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	}
}

// WithTLSConfig makes the OTLP/gRPC connection use TLS. Without it the connection is
// plaintext, which is only appropriate for a collector on the same host or pod.
func WithTLSConfig(cfg *tls.Config) SetupOption {
	return func(c *setupConfig) { c.tlsConfig = cfg }
}

// WithServiceVersion sets the service.version resource attribute (your release or
// commit), so traces can be compared before and after a deploy.
func WithServiceVersion(v string) SetupOption {
	return func(c *setupConfig) { c.version = v }
}

// WithEnvironment sets the deployment.environment resource attribute.
func WithEnvironment(env string) SetupOption {
	return func(c *setupConfig) { c.environment = env }
}

// WithResourceAttributes adds extra resource attributes (e.g. the consumer group).
func WithResourceAttributes(attrs ...attribute.KeyValue) SetupOption {
	return func(c *setupConfig) { c.attrs = append(c.attrs, attrs...) }
}

// WithBatchOptions overrides the span batcher settings (queue size, batch size, batch
// and export timeouts). They are applied after, and so win over, the defaults.
func WithBatchOptions(opts ...sdktrace.BatchSpanProcessorOption) SetupOption {
	return func(c *setupConfig) { c.batchOpts = append(c.batchOpts, opts...) }
}

func (c *setupConfig) sdkSampler() sdktrace.Sampler {
	if c.sampler != nil {
		return c.sampler
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(DefaultSampleRatio))
}

// buildResource assembles the resource describing this process: service.name, the
// optional version/environment/extra attributes, the host name, and anything in the
// standard OTEL_RESOURCE_ATTRIBUTES / OTEL_SERVICE_NAME environment variables (the
// usual way to inject k8s.pod.name and k8s.namespace.name via the downward API).
//
// It deliberately has no schema URL, so it is not pinned to one semconv version.
func (c *setupConfig) buildResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{attribute.String("service.name", serviceName)}
	if c.version != "" {
		attrs = append(attrs, attribute.String("service.version", c.version))
	}
	if c.environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment", c.environment))
	}
	attrs = append(attrs, c.attrs...)
	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithHost(),
		resource.WithFromEnv(),
	)
}

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
// Sampling defaults to ParentBased(TraceIDRatioBased(DefaultSampleRatio)); use
// WithSampleRatio or WithSampler to change it. For production, pair head sampling with
// tail sampling in the collector so errors, panics and retries are always kept.
//
// The connection is plaintext unless WithTLSConfig is given. Add WithServiceVersion and
// WithEnvironment so traces carry what is needed to compare releases.
//
// The returned shutdown func flushes any buffered spans and closes the exporter
// connection; call it during graceful shutdown (e.g. deferred in main, or alongside
// whatever already handles SIGTERM for the consumer process). It is always safe to call,
// even when enabled is false.
func Setup(ctx context.Context, serviceName, otlpEndpoint string, enabled bool, opts ...SetupOption) (shutdown func(context.Context) error, err error) {
	noopShutdown := func(context.Context) error { return nil }

	if !enabled {
		return noopShutdown, nil
	}

	var cfg setupConfig
	for _, o := range opts {
		o(&cfg)
	}

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithTimeout(defaultExportTimeout),
	}
	if cfg.tlsConfig != nil {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(cfg.tlsConfig)))
	} else {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}
	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("otelemetry: failed to create OTLP exporter: %w", err)
	}

	res, err := cfg.buildResource(ctx, serviceName)
	if err != nil && res == nil {
		return nil, fmt.Errorf("otelemetry: failed to build resource: %w", err)
	}

	batchOpts := append([]sdktrace.BatchSpanProcessorOption{
		sdktrace.WithMaxQueueSize(defaultMaxQueueSize),
		sdktrace.WithMaxExportBatchSize(defaultMaxExportBatchSize),
		sdktrace.WithBatchTimeout(defaultBatchTimeout),
		sdktrace.WithExportTimeout(defaultExportTimeout),
	}, cfg.batchOpts...)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, batchOpts...),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(cfg.sdkSampler()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
