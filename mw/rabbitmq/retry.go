package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gojekfarm/ziggurat"
	zl "github.com/gojekfarm/ziggurat/logger"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

var ErrPublisherNotInit = errors.New("auto retry publish error: publisher not initialized, please call the InitPublisher method")
var ErrCleanShutdown = errors.New("clean shutdown of rabbitmq streams")

type dsViewResp struct {
	Events []*ziggurat.Event `json:"events"`
	Count  int               `json:"count"`
}

type dsReplayResp struct {
	ReplayCount int `json:"replay_count"`
	ErrorCount  int `json:"error_count"`
}

type autoRetry struct {
	publishDialer *amqpextra.Dialer
	consumeDialer *amqpextra.Dialer
	hosts         []string
	amqpURLs      []string
	username      string
	password      string
	logger        logger.Logger
	connTimeout   time.Duration
	ogLogger      ziggurat.StructuredLogger
	queueConfig   map[string]QueueConfig
}

func constructAMQPURL(host, username, password string) string {
	return fmt.Sprintf("amqp://%s:%s@%s", username, password, host)
}

func AutoRetry(qc []QueueConfig, opts ...Opts) *autoRetry {
	r := &autoRetry{
		publishDialer: nil,
		hosts:         []string{"localhost:5672"},
		username:      "guest",
		ogLogger:      zl.NOOP,
		connTimeout:   30 * time.Second,
		password:      "guest",
		logger:        logger.Discard,
		queueConfig:   map[string]QueueConfig{},
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

func (r *autoRetry) publish(ctx context.Context, event *ziggurat.Event, queue string) error {
	if r.publishDialer == nil {
		return ErrPublisherNotInit
	}
	pub, err := getPublisher(ctx, r.publishDialer, r.logger)
	if err != nil {
		return err
	}
	defer pub.Close()
	return publishInternal(pub, queue, r.queueConfig[queue].RetryCount, r.queueConfig[queue].DelayExpirationInMS, event)
}

func (r *autoRetry) Publish(ctx context.Context, event *ziggurat.Event, queue string, queueType string, expirationInMS string) error {
	if r.publishDialer == nil {
		return ErrPublisherNotInit
	}
	exchange := fmt.Sprintf("%s_%s_%s", queue, queueType, "exchange")
	p, err := getPublisher(ctx, r.publishDialer, r.logger)
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
	defer p.Close()
	return p.Publish(msg)
}

func (r *autoRetry) Wrap(f ziggurat.HandlerFunc, queue string) ziggurat.HandlerFunc {
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

func (r *autoRetry) InitPublishers(ctx context.Context) error {

	dialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.publishDialer = dialer

	ch, err := getChannelFromDialer(ctx, r.publishDialer, r.connTimeout)
	if err != nil {
		return err
	}

	for _, qc := range r.queueConfig {
		if err := createQueuesAndExchanges(ch, qc.QueueName, r.ogLogger); err != nil {
			r.ogLogger.Error("error creating queues and exchanges", err)
			return fmt.Errorf("error iniitializing publishers:%w", err)
		}
	}
	err = ch.Close()
	r.ogLogger.Error("error closing channel", err)
	return nil
}

func (r *autoRetry) Stream(ctx context.Context, h ziggurat.Handler) error {
	dialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.consumeDialer = dialer

	ch, err := getChannelFromDialer(ctx, dialer, r.connTimeout)
	if err != nil {
		return err
	}

	for _, qc := range r.queueConfig {
		if err := createQueuesAndExchanges(ch, qc.QueueName, r.ogLogger); err != nil {
			r.ogLogger.Error("error creating queues and exchanges", err)
			return fmt.Errorf("error iniitializing publishers:%w", err)
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
		<-r.publishDialer.NotifyClosed()
		r.ogLogger.Info("stopped publisher dialer")
		<-r.publishDialer.NotifyClosed()
		r.ogLogger.Info("stopped consumer dialer")
		done <- struct{}{}
	}()

	r.publishDialer.Close()
	r.consumeDialer.Close()

	<-done
	return ErrCleanShutdown
}

func (r *autoRetry) view(ctx context.Context, queue string, count int, ack bool) ([]*ziggurat.Event, error) {
	if count < 1 {
		return []*ziggurat.Event{}, nil
	}

	d, err := newDialer(ctx, r.amqpURLs, r.logger)
	defer d.Close()
	if err != nil {
		return nil, err
	}
	ch, err := getChannelFromDialer(ctx, d, r.connTimeout)

	if err != nil {
		return nil, err
	}

	actualCount := count

	qn := fmt.Sprintf("%s_%s_%s", queue, "dlq", "queue")
	q, err := ch.QueueInspect(qn)
	if err != nil {
		return []*ziggurat.Event{}, nil
	}

	if actualCount > q.Messages {
		actualCount = q.Messages
	}
	events := make([]*ziggurat.Event, actualCount)
	for i := 0; i < actualCount; i++ {

		msg, _, err := ch.Get(qn, false)
		if err != nil {
			return []*ziggurat.Event{}, err
		}
		b := msg.Body
		var e ziggurat.Event
		err = json.Unmarshal(b, &e)
		if err != nil {
			return []*ziggurat.Event{}, err
		}

		var ackErr error
		if ack {
			ackErr = msg.Ack(true)
		} else {
			ackErr = msg.Reject(true)
		}

		r.ogLogger.Error("", ackErr)
		events[i] = &e
	}
	r.ogLogger.Error("auto retry view: channel close error:", ch.Close())
	return events, nil
}

/*DSViewHandler allows you to peek into
  the rabbitMQ dead-set queue.
*/
func (r *autoRetry) DSViewHandler(ctx context.Context) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		qname, count, err := validateQueryParams(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		events, err := r.view(ctx, qname, count, false)
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't view messages from dlq: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		je := json.NewEncoder(w)
		resp := dsViewResp{
			Events: events,
			Count:  len(events),
		}
		err = je.Encode(resp)

		if err != nil {
			http.Error(w, "json encode error", http.StatusInternalServerError)
		}
	}
	return http.HandlerFunc(f)
}

func (r *autoRetry) DeleteQueuesAndExchanges(ctx context.Context, queueName string) error {
	d, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return fmt.Errorf("error getting dialer:%w", err)
	}
	ch, err := getChannelFromDialer(ctx, d, r.connTimeout)
	if err != nil {
		return fmt.Errorf("error getting channel:%w", err)
	}

	err = deleteQueuesAndExchanges(ch, queueName)
	r.ogLogger.Error("auto retry exchange and queue deletion error", ch.Close())
	d.Close()
	return err
}

func (r *autoRetry) replay(ctx context.Context, queue string, count int) (int, error) {
	var replayCount int

	actualCount := count
	d, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return replayCount, err
	}

	ch, err := getChannelFromDialer(ctx, d, r.connTimeout)
	if err != nil {
		return replayCount, fmt.Errorf("error getting channel:%w", err)
	}

	srcQueue := fmt.Sprintf("%s_%s_%s", queue, "dlq", "queue")
	q, err := ch.QueueInspect(srcQueue)
	if err != nil {
		return replayCount, fmt.Errorf("error inspecting queue:%w", err)
	}

	if actualCount > q.Messages {
		actualCount = q.Messages
	}

	p, err := d.Publisher()
	if err != nil {
		return replayCount, fmt.Errorf("error getting publiser:%w", err)
	}
	for i := 0; i < actualCount; i++ {
		m, _, err := ch.Get(srcQueue, false)
		if err != nil {
			return replayCount, fmt.Errorf("error getting message from queue:%w", err)
		}
		err = p.Publish(publisher.Message{
			Exchange: fmt.Sprintf("%s_%s_%s", queue, "instant", "exchange"),
			Publishing: amqp.Publishing{
				Body: m.Body,
			},
		})
		if err != nil {
			return replayCount, fmt.Errorf("error publishing to instant queue:%w", err)
		}
		err = m.Ack(false)
		if err != nil {
			return replayCount, fmt.Errorf("error ack-ing maessage:%w", err)
		}
		replayCount++
	}
	return replayCount, nil
}

func (r *autoRetry) DSReplayHandler(ctx context.Context) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		qname, count, err := validateQueryParams(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		replayCount, err := r.replay(ctx, qname, count)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		je := json.NewEncoder(w)
		err = je.Encode(dsReplayResp{
			ReplayCount: replayCount,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
	return http.HandlerFunc(f)
}
