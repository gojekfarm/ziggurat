package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
)

type retry struct {
	dialer        *amqpextra.Dialer
	consumeDialer *amqpextra.Dialer
	hosts         []string
	amqpURLs      []string
	username      string
	password      string
	logger        logger.Logger
	queueConfig   map[string]QueueConfig
}

func constructAMQPURL(host, username, password string) string {
	return fmt.Sprintf("amqp://%s:%s@%s", username, password, host)
}

func AutoRetry(opts ...Opts) *retry {
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
		r.amqpURLs = append(AMQPURLs, constructAMQPURL(h, r.username, r.password))
	}
	return r
}

func (r *retry) Publish(ctx context.Context, event *ziggurat.Event, queue string) error {

	pub, err := getPublisher(ctx, r.dialer, r.logger)

	if err != nil {
		return err
	}
	defer pub.Close()
	return publish(pub, queue, r.queueConfig[queue].RetryCount, r.queueConfig[queue].DelayExpirationInMS, event)

}

func (r *retry) Wrap(f ziggurat.HandlerFunc, queue string) ziggurat.HandlerFunc {
	hf := func(ctx context.Context, event *ziggurat.Event) error {
		err := f(ctx, event)
		if err == ziggurat.Retry {
			return r.Publish(ctx, event, queue)
		}
		return nil
	}
	return hf
}

func (r *retry) Run(ctx context.Context, h ziggurat.Handler, opts ...Opts) error {

	for _, o := range opts {
		o(r)
	}

	pdialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.dialer = pdialer

	cdialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.consumeDialer = cdialer

	ch, err := getChannelFromDialer(ctx, r.dialer)
	if err != nil {
		return err
	}

	for _, qc := range r.queueConfig {
		if err := createQueuesAndExchanges(ch, qc.QueueName); err != nil {
			r.logger.Printf("error creating queues and exchanges: %v", err)
			return err
		}
	}
	ch.Close()

	consStopCh := make(chan struct{})
	for _, qc := range r.queueConfig {
		go func(qname string, wc int) {
			cons, err := startConsumer(ctx, r.consumeDialer, qname, wc, h, r.logger)
			if err != nil {
				r.logger.Printf("error starting consumer: %v", err)
			}
			<-cons.NotifyClosed()
			consStopCh <- struct{}{}
			r.logger.Printf("shutting down consumer for %s", qname)
		}(qc.QueueName, qc.WorkerCount)
	}

	for i := 0; i < len(r.queueConfig); i++ {
		<-consStopCh
	}
	close(consStopCh)

	r.dialer.Close()
	r.consumeDialer.Close()

	var closeCount int
	for {
		if closeCount == 2 {
			return nil
		}
		select {
		case <-r.consumeDialer.NotifyClosed():
			r.logger.Printf("shutting down consumer dialer")
			closeCount++
		case <-r.dialer.NotifyClosed():
			r.logger.Printf("shutting down publisher dialer")
			closeCount++
		}
	}

}
