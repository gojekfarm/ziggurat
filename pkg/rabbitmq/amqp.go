package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

var withChannel = func(connection *amqp.Connection, cb func(c *amqp.Channel) error) error {
	c, err := connection.Channel()
	defer c.Close()
	if err != nil {
		return err
	}
	cberr := cb(c)
	return cberr
}

var createDialer = func(ctx context.Context, hosts []string) (*amqpextra.Dialer, error) {
	d, cfgErr := amqpextra.NewDialer(
		amqpextra.WithURL(hosts...),
		amqpextra.WithLogger(logger.Func(func(format string, v ...interface{}) {
			msg := fmt.Sprintf(format, v...)
			zlogger.LogDebug(msg, nil)
		})),
		amqpextra.WithContext(ctx))
	if cfgErr != nil {
		return nil, cfgErr
	}
	return d, nil
}

var getConnectionFromDialer = func(ctx context.Context, d *amqpextra.Dialer, timeout time.Duration) (*amqp.Connection, error) {
	connCtx, cancelFunc := context.WithTimeout(ctx, timeout)
	conn, err := d.Connection(connCtx)
	if err != nil {
		cancelFunc()
	}
	return conn, err
}

var createPublisher = func(ctx context.Context, d *amqpextra.Dialer) (*publisher.Publisher, error) {
	options := []publisher.Option{publisher.WithContext(ctx)}
	return d.Publisher(options...)
}

var createConsumer = func(app z.App, d *amqpextra.Dialer, ctag string, queueName string, msgHandler z.MessageHandler) (*consumer.Consumer, error) {
	options := []consumer.Option{
		consumer.WithInitFunc(func(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
			channel, err := conn.(*amqp.Connection).Channel()
			if err != nil {
				return nil, err
			}
			zlogger.LogError(channel.Qos(1, 0, false), "rmq consumer: error setting QOS", nil)
			return channel, nil
		}),
		consumer.WithContext(app.Context()),
		consumer.WithConsumeArgs(ctag, false, false, false, false, nil),
		consumer.WithQueue(queueName),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			zlogger.LogInfo("rmq consumer: processing message", map[string]interface{}{"queue-name": queueName})
			msgEvent, err := decodeMessage(msg.Body)
			if err != nil {
				return msg.Reject(true)
			}
			msgHandler.HandleMessage(msgEvent, app)
			return msg.Ack(false)
		}))}
	return d.Consumer(options...)
}

var declareExchange = func(c *amqp.Channel, exchangeName string) {
	c.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil)
}

var queueBind = func(c *amqp.Channel, queueName string, exchangeName string, args amqp.Table) error {
	return c.QueueBind(queueName, "", exchangeName, false, args)
}

var queueDeclare = func(c *amqp.Channel, queueName string, args amqp.Table) (amqp.Queue, error) {
	return c.QueueDeclare(queueName, true, false, false, false, args)
}
