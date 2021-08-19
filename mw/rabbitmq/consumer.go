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

func startConsumer(ctx context.Context, d *amqpextra.Dialer, queue string, workerCount int, h ziggurat.Handler, l logger.Logger, ogl ziggurat.StructuredLogger) (*consumer.Consumer, error) {
	parallelWorker := consumer.NewParallelWorker(workerCount)
	qname := fmt.Sprintf("%s_%s_%s", queue, "instant", "queue")
	consumerName := fmt.Sprintf("%s_consumer", queue)
	cons, err := d.Consumer(
		consumer.WithContext(ctx),
		consumer.WithConsumeArgs(consumerName, false, false, false, false, amqp.Table{}),
		consumer.WithQueue(qname),
		consumer.WithQos(1, false),
		consumer.WithWorker(parallelWorker),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			bb := msg.Body
			var event ziggurat.Event
			err := json.Unmarshal(bb, &event)
			if err != nil {
				rejectErr := msg.Reject(true)
				ogl.Error("error rejecting message:", rejectErr)
			}
			ogl.Info("amqp processing message", map[string]interface{}{"consumer": consumerName}, event.Metadata)
			err = h.Handle(ctx, &event)
			if err != nil {
				ogl.Error("error processing message", err)
			}
			//return msg.Ack(false)
			return nil
		})),
	)
	if err != nil {
		return nil, err
	}
	return cons, nil
}
