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

func consumerInitFunc(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
	ch, err := conn.(*amqp.Connection).Channel()
	if err != nil {
		return nil, err
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func startConsumer(ctx context.Context, d *amqpextra.Dialer, queue string, workerCount int, h ziggurat.Handler, l logger.Logger) (*consumer.Consumer, error) {
	parallelWorker := consumer.NewParallelWorker(workerCount)
	qname := fmt.Sprintf("%s_%s_%s", queue, "instant", "queue")
	cons, err := d.Consumer(
		consumer.WithContext(ctx),
		consumer.WithQueue(qname),
		consumer.WithInitFunc(consumerInitFunc),
		consumer.WithWorker(parallelWorker),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			bb := msg.Body
			var event ziggurat.Event
			err := json.Unmarshal(bb, &event)
			if err != nil {
				rejectErr := msg.Reject(true)
				l.Printf("error rejecting message: %v", rejectErr)
			}
			err = h.Handle(ctx, &event)
			if err != nil {
				l.Printf("error processing message %v: ", err)
			}
			return msg.Ack(false)
		})),
	)
	if err != nil {
		return nil, err
	}
	return cons, nil
}
