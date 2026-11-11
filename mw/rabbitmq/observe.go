package rabbitmq

import (
	"context"
	"strings"
	"time"

	"github.com/gojekfarm/ziggurat/v2/mw/otelemetry"
	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"

	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Values of the ziggurat.retry.outcome span attribute.
const (
	outcomeRetry   = "retry"
	outcomeDLQ     = "dlq"
	outcomePublish = "publish"
)

// publishErrorReason classifies a publish failure for ziggurat_go_retry_publish_errors_total.
// amqpextra reports a broker negative acknowledgement (publisher confirms) as an error
// whose text contains "nack"; everything else is a plain publish failure.
func publishErrorReason(err error) string {
	if strings.Contains(strings.ToLower(err.Error()), "nack") {
		return prometheus.PublishErrNack
	}
	return prometheus.PublishErrPublish
}

// traceHeaders builds the outgoing AMQP headers: the retry-origin marker plus the trace
// context active on ctx (a no-op when tracing is not configured).
func traceHeaders(ctx context.Context) map[string]interface{} {
	carrier := otelemetry.HeaderCarrier{"retry-origin": "ziggurat-go"}
	otelemetry.InjectTraceContext(ctx, carrier)
	headers := make(map[string]interface{}, len(carrier))
	for k, v := range carrier {
		headers[k] = v
	}
	return headers
}

// publishObserved publishes msg and records what happened: a PRODUCER span (whose
// context is what gets injected into the message headers, so the next consumer is a
// child of it), the publish latency, and on failure the error counter with its reason.
//
// outcome is one of outcomeRetry/outcomeDLQ/outcomePublish and retryCount the retry
// count the message carries after this publish.
func publishObserved(ctx context.Context, p amqpPublisher, queue, outcome string, retryCount int, msg publisher.Message) error {
	ctx, span := otelemetry.StartPublishSpan(ctx, queue,
		attribute.String("ziggurat.retry.outcome", outcome),
		attribute.Int("ziggurat.retry.count", retryCount),
	)
	defer span.End()

	msg.Publishing.Headers = traceHeaders(ctx)

	start := time.Now()
	err := p.Publish(msg)
	elapsed := time.Since(start).Seconds()

	result := "success"
	if err != nil {
		result = "error"
		reason := publishErrorReason(err)
		prometheus.RetryPublishErrorsCounter.WithLabelValues(queue, reason).Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, reason)
	}
	prometheus.RetryPublishDurationHistogram.WithLabelValues(queue, result).Observe(elapsed)
	return err
}

// replayHeaders copies the original message's headers and adds the replay marker, without
// mutating the delivery's own header table.
func replayHeaders(orig amqp.Table) amqp.Table {
	h := make(amqp.Table, len(orig)+1)
	for k, v := range orig {
		h[k] = v
	}
	h[otelemetry.HeaderReplay] = "true"
	return h
}
