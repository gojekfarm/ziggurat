package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
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

	dialer, err := newDialer(ctx, AMQPURLs, r.logger)
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

	err = createQueuesAndExchanges(ch, pubOpts.queueKey)
	if err != nil {
		return err
	}

	err = ch.Close()
	if err != nil {
		r.logger.Printf("error closing channel: " + err.Error())
	}
	pub, err := getPublisher(ctx, r.dialer, r.logger)

	if err != nil {
		return err
	}
	defer pub.Close()
	return publish(pub, pubOpts.queueKey, pubOpts.retryCount, pubOpts.delayExpiration, event)

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
