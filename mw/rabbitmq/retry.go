package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"
	"net/http"
	"net/url"
	"sync"
	"time"

	zl "github.com/gojekfarm/ziggurat/v2/logger"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

var ErrPublisherNotInit = errors.New("auto retry publish error: publisher not initialized, please call the InitPublisher method")
var ErrCleanShutdown = errors.New("clean shutdown of rabbitmq streams")

const (
	QueueTypeDL      = "dlq"
	QueueTypeInstant = "instant"
	QueueTypeDelay   = "delay"
)

type dsViewResp struct {
	Events []*ziggurat.Event `json:"events"`
	Count  int               `json:"count"`
}

type dsReplayResp struct {
	ReplayCount int `json:"replay_count"`
	ErrorCount  int `json:"error_count"`
}

type ARetry struct {
	publishDialer *amqpextra.Dialer
	consumeDialer *amqpextra.Dialer
	once          sync.Once
	hosts         []string
	amqpURLs      []string
	username      string
	password      string
	logger        logger.Logger
	connTimeout   time.Duration
	ogLogger      ziggurat.StructuredLogger
	queueConfig   map[string]QueueConfig
	publisherPool *publisherPool
}

func constructAMQPURL(host, username, password string) string {
	escapedUser := url.QueryEscape(username)
	escapedPass := url.QueryEscape(password)
	return fmt.Sprintf("amqp://%s:%s@%s", escapedUser, escapedPass, host)
}

func AutoRetry(qc Queues, opts ...Opts) *ARetry {
	r := &ARetry{
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

		if c.ConsumerCount < 1 {
			c.ConsumerCount = 1
		}
		r.queueConfig[c.QueueKey] = c
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

func (r *ARetry) publish(c context.Context, event *ziggurat.Event, queue string) error {
	if r.publishDialer == nil {
		return ErrPublisherNotInit
	}

	pub, err := r.publisherPool.get(c)
	if err != nil {
		return err
	}
	defer r.publisherPool.put(pub)
	err = publishInternal(pub, queue, r.queueConfig[queue].RetryCount, r.queueConfig[queue].DelayExpirationInMS, event)
	return err
}

// Publish can be called from anywhere and messages can be sent to any queue
func (r *ARetry) Publish(ctx context.Context, event *ziggurat.Event, queueKey string, queueType string, expirationMS string) error {
	r.once.Do(func() {
		r.ogLogger.Info("[amqp] init from publish")
		err := r.InitPublishers(ctx)
		if err != nil {
			panic(fmt.Sprintf("could not start RabbitMQ publishers:%v", err))
		}
	})
	var err error
	exchange := fmt.Sprintf("%s_%s", queueKey, "exchange")
	p, err := r.publisherPool.get(ctx)
	if err != nil {
		return err
	}
	defer r.publisherPool.put(p)
	if err != nil {
		return err
	}
	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := publisher.Message{
		Exchange: exchange,
		Key:      queueType,
		Publishing: amqp.Publishing{
			Expiration: expirationMS,
			Body:       eb,
			Headers:    map[string]interface{}{"retry-origin": "ziggurat-go"},
		},
	}
	return p.Publish(msg)
}

func (r *ARetry) Retry(ctx context.Context, event *ziggurat.Event, queueKey string) error {
	r.once.Do(func() {
		r.ogLogger.Info("[amqp] running init function from retry")
		err := r.InitPublishers(ctx)
		if err != nil {
			panic(fmt.Sprintf("could not start RabbitMQ publishers:%v", err))
		}
	})
	return r.publish(ctx, event, queueKey)
}

func (r *ARetry) InitPublishers(ctx context.Context) error {
	dialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.publishDialer = dialer
	r.publisherPool, err = newPubPool(10, dialer, r.logger)
	if err != nil {
		return err
	}

	ch, err := getChannelFromDialer(ctx, r.publishDialer, r.connTimeout)
	if err != nil {
		return err
	}

	for _, qc := range r.queueConfig {
		if err := createQueuesAndExchanges(ch, qc.QueueKey, r.ogLogger); err != nil {
			r.ogLogger.Error("error creating queues and exchanges", err)
			return fmt.Errorf("error iniitializing publishers:%w", err)
		}
	}
	err = ch.Close()
	r.ogLogger.Error("error closing channel", err)
	return nil
}

func (r *ARetry) Consume(ctx context.Context, h ziggurat.Handler) error {
	dialer, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return err
	}
	r.consumeDialer = dialer

	ch, err := getChannelFromDialer(ctx, dialer, r.connTimeout)
	if err != nil {
		return err
	}

	// twice called
	for _, qc := range r.queueConfig {
		if err := createQueuesAndExchanges(ch, qc.QueueKey, r.ogLogger); err != nil {
			r.ogLogger.Error("error creating queues and exchanges", err)
			return fmt.Errorf("error iniitializing publishers:%w", err)
		}
	}
	err = ch.Close()
	r.ogLogger.Error("error closing channel", err)

	var wg sync.WaitGroup
	for _, qc := range r.queueConfig {
		for i := 0; i < qc.ConsumerCount; i++ {
			wg.Add(1)
			go func(qc QueueConfig) {
				cons, err := startConsumer(ctx, r.consumeDialer, qc, h, r.logger, r.ogLogger)
				if err != nil {
					r.ogLogger.Error("error starting consumer", err)
				}
				<-cons.NotifyClosed()
				wg.Done()
			}(qc)
		}
	}

	wg.Wait()

	return ErrCleanShutdown
}

func (r *ARetry) view(ctx context.Context, qnameWithType string, count int, ack bool) ([]*ziggurat.Event, error) {
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

	q, err := ch.QueueInspect(qnameWithType)
	if err != nil {
		return []*ziggurat.Event{}, nil
	}

	defer ch.Close()

	if actualCount > q.Messages {
		actualCount = q.Messages
	}
	events := make([]*ziggurat.Event, actualCount)
	for i := 0; i < actualCount; i++ {

		msg, _, err := ch.Get(qnameWithType, false)
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
		}

		r.ogLogger.Error("", ackErr)
		events[i] = &e
	}
	r.ogLogger.Error("auto retry view: channel close error:", ch.Close())
	return events, nil
}

/*
DSViewHandler allows you to peek into

	the rabbitMQ dead-set queue.
*/
func (r *ARetry) DSViewHandler(ctx context.Context) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		qname, count, err := validateQueryParams(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		qn := fmt.Sprintf("%s_%s_%s", qname, QueueTypeDL, "queue")
		events, err := r.view(ctx, qn, count, false)
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

func (r *ARetry) DeleteQueuesAndExchanges(ctx context.Context, queueName string) error {
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

func (r *ARetry) replay(ctx context.Context, queue string, count int) (int, error) {
	var replayCount int

	actualCount := count
	d, err := newDialer(ctx, r.amqpURLs, r.logger)
	if err != nil {
		return replayCount, err
	}
	defer d.Close()
	ch, err := getChannelFromDialer(ctx, d, r.connTimeout)
	if err != nil {
		return replayCount, fmt.Errorf("error getting channel:%w", err)
	}

	defer func() {
		if err := ch.Close(); err != nil {
			r.ogLogger.Error("error closing channel", err)
		}
	}()

	srcQueue := fmt.Sprintf("%s_%s_%s", queue, QueueTypeDL, "queue")
	q, err := ch.QueueInspect(srcQueue)
	if err != nil {
		return replayCount, fmt.Errorf("error inspecting queue:%w", err)
	}

	if actualCount > q.Messages {
		actualCount = q.Messages
	}

	p, err := d.Publisher()
	defer p.Close()
	if err != nil {
		return replayCount, fmt.Errorf("error getting publiser:%w", err)
	}
	for i := 0; i < actualCount; i++ {
		m, _, err := ch.Get(srcQueue, false)
		if err != nil {
			return replayCount, fmt.Errorf("error getting message from queue:%w", err)
		}
		err = p.Publish(publisher.Message{
			Key:      QueueTypeInstant,
			Exchange: fmt.Sprintf("%s_%s", queue, "exchange"),
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

func (r *ARetry) DSReplayHandler(ctx context.Context) http.Handler {
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
