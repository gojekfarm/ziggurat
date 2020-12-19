package retry

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/streadway/amqp"
)

var decodeMessage = func(body []byte) (*ziggurat.Message, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := ziggurat.NewMessage(nil, nil, "", map[string]interface{}{})
	if decodeErr := decoder.Decode(messageEvent); decodeErr != nil {
		return messageEvent, decodeErr
	}
	return messageEvent, nil
}

var createConsumer = func(ctx context.Context, d *amqpextra.Dialer, ctag string, queueName string, msgHandler ziggurat.Handler, l ziggurat.LeveledLogger) (*consumer.Consumer, error) {
	options := []consumer.Option{
		consumer.WithInitFunc(func(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
			channel, err := conn.(*amqp.Connection).Channel()
			if err != nil {
				return nil, err
			}
			l.Errorf("rabbitmq: error setting QOS, %v", channel.Qos(1, 0, false))
			return channel, nil
		}),
		consumer.WithContext(ctx),
		consumer.WithConsumeArgs(ctag, false, false, false, false, nil),
		consumer.WithQueue(queueName),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			l.Infof("rabbitmq processing message from QUEUE_NAME %s", queueName)
			msgEvent, err := decodeMessage(msg.Body)
			if err != nil {
				return msg.Reject(true)
			}
			msgHandler.HandleMessage(msgEvent, ctx)
			return msg.Ack(false)
		}))}
	return d.Consumer(options...)
}
