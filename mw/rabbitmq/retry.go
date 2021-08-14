package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

type retry struct {
	dialer   *amqpextra.Dialer
	hosts    []string
	username string
	password string
	logger   logger.Logger
}

func constructAMQPURL(host, username, password string) string {
	return fmt.Sprintf("amqp://%s:%s@%s", username, password, host)
}

func NewRetry(ctx context.Context, opts ...Opts) (*retry, error) {
	r := &retry{
		dialer:   nil,
		hosts:    []string{"localhost:5672"},
		username: "guest",
		password: "guest",
		logger:   logger.Discard,
	}
	for _, o := range opts {
		o(r)
	}
	AMQPURLs := make([]string, 0, len(r.hosts))
	for _, h := range r.hosts {
		AMQPURLs = append(AMQPURLs, constructAMQPURL(h, r.username, r.password))
	}

	dialer, err := NewDialer(ctx, AMQPURLs, r.logger)
	if err != nil {
		return nil, err
	}
	r.dialer = dialer
	return r, nil
}

func (r *retry) Publish(ctx context.Context, event *ziggurat.Event, opts ...PublishOpts) error {

	pubOpts := publishOpts{
		retryCount:      1,
		queueKey:        event.Path,
		delayExpiration: "2000",
	}

	for _, o := range opts {
		o(&pubOpts)
	}

	ch, err := getChannelFromDialer(ctx, r.dialer)
	if err != nil {
		return err
	}

	queueTypes := []string{"delay", "instant", "dlq"}
	for _, qt := range queueTypes {
		args := amqp.Table{}
		queueName := fmt.Sprintf("%s_%s", pubOpts.queueKey, qt)
		if qt == "delay" {
			args = amqp.Table{"x-dead-letter-exchange": fmt.Sprintf("%s_%s_%s", pubOpts.queueKey, "instant", "exchange")}
		}
		if err := CreateAndBindQueue(ch, queueName, args); err != nil {
			return err
		}
	}

	pub, err := r.dialer.Publisher(
		publisher.WithContext(ctx),
		publisher.WithLogger(r.logger))

	if err != nil {
		return err
	}
	defer pub.Close()
	newCount := getRetryCount(event) + 1

	if newCount > pubOpts.retryCount {
		fmt.Printf("retry count %d: publishing to dead letter", newCount)
		return nil
	}

	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	exchange := fmt.Sprintf("%s_%s_%s", pubOpts.queueKey, "delay", "exchange")
	msg := publisher.Message{
		Exchange: exchange,
		Publishing: amqp.Publishing{
			Expiration: pubOpts.delayExpiration,
			Body:       eb,
		},
	}

	err = pub.Publish(msg)
	return err
}

func (r *retry) Wrap(f ziggurat.HandlerFunc, opts ...PublishOpts) ziggurat.HandlerFunc {
	hf := func(ctx context.Context, event *ziggurat.Event) error {
		err := f(ctx, event)
		if err == ziggurat.Retry {
			return r.Publish(ctx, event, opts...)
		}
		return nil
	}
	return hf
}
