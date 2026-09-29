package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/gojekfarm/ziggurat/v2"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/streadway/amqp"
)

// handleDelivery contains the actual per-message logic for the instant-queue consumer:
// deserialize, invoke the handler, and ack/reject. It is a standalone function (rather
// than an inline closure) specifically so it can be unit tested without needing a live
// RabbitMQ connection — amqp.Delivery is a plain struct that can be constructed
// directly in a test with a fake Acknowledger.
//
// A panic in h.Handle is recovered here rather than allowed to propagate: unrecovered,
// it would climb up through amqpextra's own goroutine and crash the whole process,
// taking down every other queue this process happens to be consuming from.
func handleDelivery(ctx context.Context, msg amqp.Delivery, h ziggurat.Handler, ogl ziggurat.StructuredLogger, consumerName string) (ackResult interface{}) {
	bb := msg.Body
	var event ziggurat.Event
	err := json.Unmarshal(bb, &event)
	if err != nil {
		ogl.Error("amqp unmarshal error", err)
		return msg.Reject(true)
	}
	ogl.Info("amqp processing message", map[string]interface{}{"consumer": consumerName})

	defer func() {
		if r := recover(); r != nil {
			ogl.Error(
				"recovered from panic in handler",
				fmt.Errorf("panic: %v", r),
				map[string]interface{}{
					"consumer": consumerName,
					"stack":    string(debug.Stack()),
				},
			)
			// Deliberately still Ack here, matching this handler's existing
			// happy-path behavior, rather than silently changing message delivery
			// semantics as a side effect of adding panic recovery. Whether a
			// panicking message should instead be routed to the DLQ is a separate,
			// deliberate decision — see the RabbitMQ poison-message follow-up ticket.
			ackResult = msg.Ack(false)
		}
	}()

	h.Handle(ctx, &event)
	return msg.Ack(false)
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
			return handleDelivery(ctx, msg, h, ogl, consumerName)
		})),
	)

	if err != nil {
		return nil, err
	}
	return cons, nil
}
