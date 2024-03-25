package rabbitmq

import (
	"context"
	"time"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

func newDialer(ctx context.Context, AMQPURLs []string, l logger.Logger) (*amqpextra.Dialer, error) {
	dialer, err := amqpextra.NewDialer(
		amqpextra.WithContext(ctx),
		amqpextra.WithLogger(l),
		amqpextra.WithConnectionProperties(amqp.Table{"connection_name": "ziggurat_go"}),
		amqpextra.WithURL(AMQPURLs...))
	if err != nil {
		return nil, err
	}
	return dialer, nil
}

func getChannelFromDialer(ctx context.Context, d *amqpextra.Dialer, timeout time.Duration) (*amqp.Channel, error) {
	timeoutCtx, cfn := context.WithTimeout(ctx, timeout)
	done := timeoutCtx.Done()
	defer cfn()
	go func() {
		<-done
		cfn()
	}()

	conn, err := d.Connection(timeoutCtx)
	if err != nil {
		return nil, err
	}
	return conn.Channel()
}

func getPublisher(ctx context.Context, d *amqpextra.Dialer, l logger.Logger) (*publisher.Publisher, error) {
	return d.Publisher(
		publisher.WithContext(ctx),
		publisher.WithLogger(l))
}
