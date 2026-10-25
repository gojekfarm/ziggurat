package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/gojekfarm/ziggurat/v2/mw/otelemetry"
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

	if newCount > retryCount {
		event.Metadata[KeyRetryCount] = retryCount
		routingKey = QueueTypeDL
		expiration = ""
		prometheus.DLQPublishedCounter.With(map[string]string{prometheus.QueueLabel: queue}).Inc()
	} else {
		event.Metadata[KeyRetryCount] = newCount
		prometheus.RetryPublishedCounter.With(map[string]string{prometheus.QueueLabel: queue}).Inc()
	}

	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Inject the current span's trace context (from ctx, the context active when this
	// retry/DLQ publish happens - e.g. inside the handler that's calling r.Retry) into
	// the outgoing message's headers, so this republish is a child of the same trace as
	// the original consume, not an orphaned new one. See mw/otelemetry.TraceHandler,
	// which extracts this back out on the next consume.
	carrier := otelemetry.HeaderCarrier{"retry-origin": "ziggurat-go"}
	otelemetry.InjectTraceContext(ctx, carrier)

	headers := make(map[string]interface{}, len(carrier))
	for k, v := range carrier {
		headers[k] = v
	}

	msg := publisher.Message{
		Exchange: exchange,
		Key:      routingKey,
		Publishing: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Expiration:   expiration,
			Body:         eb,
			Headers:      headers,
		},
	}

	return p.Publish(msg)
}
