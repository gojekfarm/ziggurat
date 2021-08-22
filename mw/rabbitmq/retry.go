package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	zl "github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/server"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"net/http"
	"strconv"
)

type managementServerResponse struct {
	Events []*ziggurat.Event `json:"events"`
	Count  int               `json:"count"`
}

type retry struct {
	dialer           *amqpextra.Dialer
	consumeDialer    *amqpextra.Dialer
	hosts            []string
	amqpURLs         []string
	username         string
	password         string
	logger           logger.Logger
	ogLogger         ziggurat.StructuredLogger
	queueConfig      map[string]QueueConfig
	managementServer *server.DefaultHttpServer
}

func constructAMQPURL(host, username, password string) string {
	return fmt.Sprintf("amqp://%s:%s@%s", username, password, host)
}

func AutoRetry(qc []QueueConfig, opts ...Opts) *retry {
	r := &retry{
		dialer:      nil,
		hosts:       []string{"localhost:5672"},
		username:    "guest",
		ogLogger:    zl.NewDiscardLogger(),
		password:    "guest",
		logger:      logger.Discard,
		queueConfig: map[string]QueueConfig{},
	}

	for _, c := range qc {
		r.queueConfig[c.QueueName] = c
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

func (r *retry) publish(ctx context.Context, event *ziggurat.Event, queue string) error {

	pub, err := getPublisher(ctx, r.dialer, r.logger)

	if err != nil {
		return err
	}
	defer pub.Close()
	return publishInternal(pub, queue, r.queueConfig[queue].RetryCount, r.queueConfig[queue].DelayExpirationInMS, event)
}

func (r *retry) Publish(ctx context.Context, event *ziggurat.Event, queue string, queueType string, expirationInMS string) error {
	exchange := fmt.Sprintf("%s_%s_%s", queue, queueType, "exchange")
	p, err := getPublisher(ctx, r.dialer, r.logger)
	defer p.Close()
	if err != nil {
		return err
	}
	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := publisher.Message{
		Exchange: exchange,
		Publishing: amqp.Publishing{
			Expiration: expirationInMS,
			Body:       eb,
		},
	}
	return p.Publish(msg)
}

func (r *retry) Wrap(f ziggurat.HandlerFunc, queue string) ziggurat.HandlerFunc {
	hf := func(ctx context.Context, event *ziggurat.Event) error {
		err := f(ctx, event)
		if err == ziggurat.Retry {
			pubErr := r.publish(ctx, event, queue)
			r.ogLogger.Error("AR publishInternal error", pubErr)
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
		go func(qc QueueConfig) {
			cons, err := startConsumer(ctx, r.consumeDialer, qc, h, r.logger, r.ogLogger)
			if err != nil {
				r.ogLogger.Error("error starting consumer", err)
			}
			<-cons.NotifyClosed()
			consStopCh <- struct{}{}
			r.ogLogger.Info("shutting down consumer for", map[string]interface{}{"queue": qc.QueueName})
		}(qc)
	}

	for i := 0; i < len(r.queueConfig); i++ {
		<-consStopCh
	}
	close(consStopCh)

	done := make(chan struct{})

	go func() {
		<-r.dialer.NotifyClosed()
		r.ogLogger.Info("stopped publisher dialer")
		<-r.dialer.NotifyClosed()
		r.ogLogger.Info("stopped consumer dialer")
		done <- struct{}{}
	}()

	r.dialer.Close()
	r.consumeDialer.Close()

	<-done
	return nil
}

func (r *retry) view(ctx context.Context, queue string, count int) ([]*ziggurat.Event, error) {
	ch, err := getChannelFromDialer(ctx, r.dialer)

	if err != nil {
		return nil, err
	}

	deliveries := make([]*ziggurat.Event, count)
	qn := fmt.Sprintf("%s_%s_%s", queue, "dlq", "queue")
	for i := 0; i < count; i++ {
		msg, _, err := ch.Get(qn, false)
		if err != nil {
			return nil, err
		}
		b := msg.Body
		var e ziggurat.Event
		err = json.Unmarshal(b, &e)
		if err != nil {
			return nil, err
		}
		deliveries[i] = &e
	}
	return deliveries, nil
}

func (r *retry) DSViewHandler(ctx context.Context) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		qparams := req.URL.Query()
		qname, ok := qparams["queue"]
		if !ok {
			http.Error(w, "expected query param qname", 400)
			return
		}

		c, ok := qparams["count"]
		if !ok {
			http.Error(w, "expected query param count", 400)
			return
		}
		count, err := strconv.Atoi(c[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("expected count to be a number: %v", err), 400)
			return
		}
		events, err := r.view(ctx, qname[0], count)
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't view messages from dlq: %v", err), 500)
			return
		}

		je := json.NewEncoder(w)
		resp := managementServerResponse{
			Events: events,
			Count:  len(events),
		}
		err = je.Encode(resp)

		if err != nil {
			http.Error(w, "json encode error", 500)
		}
	}
	return http.HandlerFunc(f)
}
