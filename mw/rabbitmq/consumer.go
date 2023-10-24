package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/streadway/amqp"
)

func startConsumer(ctx context.Context, d *amqpextra.Dialer, c QueueConfig, h ziggurat.Handler, l logger.Logger, ogl ziggurat.StructuredLogger) (*consumer.Consumer, error) {
	pfc := 1

	if c.ConsumerPrefetchCount > 1 {
		pfc = c.ConsumerPrefetchCount
	}

	qname := fmt.Sprintf("%s_%s_%s", c.QueueKey, QueueTypeInstant, "queue")

	if c.Type == WorkerQueue {
		qname = fmt.Sprintf("%s_%s_%s", c.QueueKey, QueueTypeWorker, "queue")
	}

	consumerName := fmt.Sprintf("%s_consumer", c.QueueKey)
	rejectCorruptMessage := true
	if c.DiscardCorruptMessage {
		rejectCorruptMessage = false
	}

	cons, err := d.Consumer(
		consumer.WithContext(ctx),
		consumer.WithQueue(qname),
		consumer.WithLogger(l),
		consumer.WithQos(pfc, false),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			bb := msg.Body
			var event ziggurat.Event
			err := json.Unmarshal(bb, &event)

			if err != nil {
				ogl.Error("amqp unmarshal error", err)
				return msg.Reject(rejectCorruptMessage)
			}
			ogl.Info("amqp processing message", map[string]interface{}{"consumer": consumerName})
			err = h.Handle(ctx, &event)
			if err != nil {
				ogl.Error("amqp message processing error", err, event.Metadata)
			}
			return msg.Ack(false)
		})),
	)

	if err != nil {
		return nil, err
	}
	return cons, nil
}
