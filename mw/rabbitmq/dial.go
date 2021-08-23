package rabbitmq

import (
	"context"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

var newDialer = func(ctx context.Context, AMQPURLs []string, l logger.Logger) (*amqpextra.Dialer, error) {
	dialer, err := amqpextra.NewDialer(
		amqpextra.WithContext(ctx),
		amqpextra.WithLogger(l),
		amqpextra.WithURL(AMQPURLs...))
	if err != nil {
		return nil, err
	}
	return dialer, nil
}

var getChannelFromDialer = func(ctx context.Context, d *amqpextra.Dialer) (*amqp.Channel, error) {
	conn, err := d.Connection(ctx)
	if err != nil {
		return nil, err
	}
	return conn.Channel()
}

var getPublisher = func(ctx context.Context, d *amqpextra.Dialer, l logger.Logger) (*publisher.Publisher, error) {
	return d.Publisher(
		publisher.WithContext(ctx),
		publisher.WithLogger(l))
}
