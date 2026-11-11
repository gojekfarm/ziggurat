package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/streadway/amqp"
)

// amqpHeadersToMap converts amqp.Table (the native AMQP message headers, a
// map[string]interface{}) into a map[string]string suitable for
// Event.Metadata[ziggurat.MetadataKeyHeaders]. Non-string header values are formatted
// via fmt.Sprintf rather than dropped, since AMQP headers can carry non-string types
// (ints, bools) that are still meaningful even if this particular vehicle is primarily
// intended for string-valued headers like a W3C traceparent.
func amqpHeadersToMap(headers amqp.Table) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	m := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			m[k] = s
			continue
		}
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}

// decodeDelivery turns a delivery body into the event handed to the handler and stamps
// the metadata that says where it was just delivered from.
//
// The queue key is always (re)set: a retried event is the original event re-serialised
// with its Metadata, so without this a message coming off RabbitMQ would still look like
// it came from Kafka to metrics, tracing and logging middlewares.
func decodeDelivery(body []byte, headers amqp.Table, queue string) (*ziggurat.Event, error) {
	var event ziggurat.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	if event.Metadata == nil {
		event.Metadata = map[string]interface{}{}
	}
	event.Metadata[ziggurat.MetadataKeyRabbitMQQueue] = queue
	if h := amqpHeadersToMap(headers); h != nil {
		event.Metadata[ziggurat.MetadataKeyHeaders] = h
	}
	return &event, nil
}

func startConsumer(ctx context.Context, d *amqpextra.Dialer, c QueueConfig, h ziggurat.Handler, l logger.Logger, ogl ziggurat.StructuredLogger) (*consumer.Consumer, error) {
	pfc := 1

	if c.ConsumerPrefetchCount > 1 {
		pfc = c.ConsumerPrefetchCount
	}

	ogl.Info("starting consumer", map[string]any{"name": c.QueueKey, "count": c.ConsumerCount})
	qname := fmt.Sprintf("%s_%s_%s", c.QueueKey, QueueTypeInstant, "queue")
	consumerName := fmt.Sprintf("%s_consumer", c.QueueKey)
	cons, err := d.Consumer(
		consumer.WithContext(ctx),
		consumer.WithQueue(qname),
		consumer.WithLogger(l),
		consumer.WithQos(pfc, false),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			event, err := decodeDelivery(msg.Body, msg.Headers, qname)
			if err != nil {
				ogl.Error("amqp unmarshal error", err)
				return msg.Reject(true)
			}
			ogl.Info("amqp processing message", map[string]interface{}{"consumer": consumerName})
			h.Handle(ctx, event)
			return msg.Ack(false)
		})),
	)

	if err != nil {
		return nil, err
	}
	return cons, nil
}
