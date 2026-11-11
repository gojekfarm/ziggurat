package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"

	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

func publishInternal(ctx context.Context, p amqpPublisher, queue string, retryCount int, delayExpiration string, event *ziggurat.Event) error {

	expiration := delayExpiration

	if event.Metadata == nil {
		event.Metadata = map[string]interface{}{KeyRetryCount: 0}
	}

	newCount := RetryCountFor(event) + 1
	exchange := fmt.Sprintf("%s_%s", queue, "exchange")
	routingKey := QueueTypeDelay
	outcome := outcomeRetry
	publishedCount := newCount
	published := prometheus.RetryPublishedCounter

	if newCount > retryCount {
		event.Metadata[KeyRetryCount] = retryCount
		routingKey = QueueTypeDL
		expiration = ""
		outcome = outcomeDLQ
		publishedCount = retryCount
		published = prometheus.DLQPublishedCounter
	} else {
		event.Metadata[KeyRetryCount] = newCount
	}

	eb, err := json.Marshal(event)
	if err != nil {
		prometheus.RetryPublishErrorsCounter.WithLabelValues(queue, prometheus.PublishErrMarshal).Inc()
		return err
	}

	msg := publisher.Message{
		Exchange: exchange,
		Key:      routingKey,
		Publishing: amqp.Publishing{
			Expiration: expiration,
			Body:       eb,
		},
	}

	// The retry/DLQ publish gets its own PRODUCER span, and that span's context (not the
	// caller's) is injected into the headers, so the next consume is a child of this
	// publish, which is itself a child of the consume that triggered the retry. See
	// publishObserved and mw/otelemetry.TraceHandler, which extracts it on the next consume.
	if err := publishObserved(ctx, p, queue, outcome, publishedCount, msg); err != nil {
		return err
	}
	// Counted only after the publish succeeded: failures are in
	// ziggurat_go_retry_publish_errors_total instead.
	published.With(map[string]string{prometheus.QueueLabel: queue}).Inc()
	return nil
}
