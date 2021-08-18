package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	zl "github.com/gojekfarm/ziggurat/logger"
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
	ogLogger      ziggurat.StructuredLogger
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
		ogLogger: zl.NewDiscardLogger(),
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
			pubErr := r.Publish(ctx, event, queue)
			r.ogLogger.Error("AR publish error", pubErr)
			// return the original error
			return err
		}
		// return the original error and not nil
		return err
	}
	return hf
}

func (r *retry) InitPublishers(ctx context.Context) error {
	pdialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.dialer = pdialer
	return nil
}

func (r *retry) Stream(ctx context.Context, h ziggurat.Handler) error {
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
		if err := createQueuesAndExchanges(ch, qc.QueueName, r.ogLogger); err != nil {
			r.ogLogger.Error("error creating queues and exchanges", err)
			return err
		}
	}
	err = ch.Close()
	r.ogLogger.Error("error closing channel", err)

	consStopCh := make(chan struct{})
	for _, qc := range r.queueConfig {
		go func(qname string, wc int) {
			cons, err := startConsumer(ctx, r.consumeDialer, qname, wc, h, r.logger, r.ogLogger)
			if err != nil {
				r.ogLogger.Error("error starting consumer", err)
			}
			<-cons.NotifyClosed()
			consStopCh <- struct{}{}
			r.ogLogger.Info("shutting down consumer for", map[string]interface{}{"queue": qname})
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
			r.ogLogger.Info("shutting down consumer dialer")
			closeCount++
		case <-r.dialer.NotifyClosed():
			r.ogLogger.Info("shutting down publisher dialer")
			closeCount++
		}
	}

}
