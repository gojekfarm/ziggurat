package retry

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	amqpl "github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

func createContextWithDeadline(parentCtx context.Context, afterTimeInS int) (context.Context, context.CancelFunc) {
	deadlineTime := time.Now().Add(time.Duration(afterTimeInS) * time.Second)
	return context.WithDeadline(parentCtx, deadlineTime)
}

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
		amqpextra.WithLogger(amqpl.Func(func(format string, v ...interface{}) {
			msg := fmt.Sprintf(format, v...)
			logger.LogDebug(msg, nil)
		})),
		amqpextra.WithContext(ctx))
	if cfgErr != nil {
		return nil, cfgErr
	}
	return d, nil
}

var getConnectionFromDialer = func(ctx context.Context, d *amqpextra.Dialer, timeout time.Duration) (*amqp.Connection, error) {
	connCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(timeout*time.Second))
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
			logger.LogError(channel.Qos(1, 0, false), "rmq consumer: error setting QOS", nil)
			return channel, nil
		}),
		consumer.WithContext(app.Context()),
		consumer.WithConsumeArgs(ctag, false, false, false, false, nil),
		consumer.WithQueue(queueName),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			logger.LogInfo("rmq consumer: processing message", map[string]interface{}{"queue-name": queueName})
			msgEvent, err := decodeMessage(msg.Body)
			if err != nil {
				return msg.Reject(true)
			}
			msgHandler.HandleMessage(msgEvent, app)
			return msg.Ack(false)
		}))}
	return d.Consumer(options...)
}
