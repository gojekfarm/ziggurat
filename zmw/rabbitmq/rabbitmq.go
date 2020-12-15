package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/zretry"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

type Opts func(o *Options)

type Options struct {
	hosts       []string
	queuePrefix string
	retryCount  int
}

type RabbitMQRetry struct {
	next    ztype.MessageHandler
	pdialer *amqpextra.Dialer
	*Options
}

func RabbitMQ(ctx context.Context, opts ...Opts) func(next ztype.MessageHandler) ztype.MessageHandler {
	r := NewRabbitMQRetry(opts...)
	err := r.Connect(ctx)
	if err != nil {
		zlog.LogFatal(err, "", nil)
	}
	return func(next ztype.MessageHandler) ztype.MessageHandler {
		r.next = next
		return r
	}
}

func NewRabbitMQRetry(opts ...Opts) *RabbitMQRetry {
	o := &Options{}
	r := &RabbitMQRetry{nil, nil, o}
	for _, opt := range opts {
		opt(o)
	}
	if r.retryCount == 0 {
		r.retryCount = 1
	}
	if r.queuePrefix == "" {
		r.queuePrefix = "ziggurat_rabbitmq"
	}
	if len(r.hosts) < 1 {
		r.hosts = []string{"amqp://user:bitnami@localhost:5672/"}
	}
	return r
}

func (r *RabbitMQRetry) HandleMessage(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
	status := r.next.HandleMessage(event, app)
	if status == ztype.RetryMessage {
		if err := r.retry(app, event); err != nil {
			zlog.LogFatal(err, "ZIGGURAT RABBITMQ MW", nil)
		}
	}
	return status
}

func (r *RabbitMQRetry) Connect(ctx context.Context) error {
	dialer, err := zretry.CreateDialer(ctx, r.hosts)
	if err != nil {
		return err
	}
	r.pdialer = dialer
	return nil
}

func (r *RabbitMQRetry) retry(app ztype.App, event zbase.MessageEvent) error {
	ctx := app.Context()
	p, err := createPublisher(ctx, r.pdialer)
	if err != nil {
		return fmt.Errorf("error creating publisher: %s", err.Error())
	}
	defer p.Close()
	return r.retryWithTimeout(ctx, event)
}

func (r *RabbitMQRetry) getQueueTypeAndExpiration(event zbase.MessageEvent) (string, string) {
	currentCount := getRetryCount(&event)
	if currentCount >= r.retryCount {
		return "dead_letter", ""
	}
	return "instant", "2000"

}

func (r *RabbitMQRetry) retryWithTimeout(ctx context.Context, payload zbase.MessageEvent) error {
	retryCountPayload := zretry.GetRetryCount(&payload)
	conn, err := zretry.GetConnectionFromDialer(ctx, r.pdialer, 30*time.Second)
	if err != nil {
		return err
	}
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	p, err := createPublisher(ctx, r.pdialer)

	queueType, expiration := r.getQueueTypeAndExpiration(payload)
	queueName := constructQueueName(r.queuePrefix, payload.StreamRoute, queueType)
	exchangeName := constructExchangeName(r.queuePrefix, payload.StreamRoute, queueType)

	delayQueueName := constructQueueName(r.queuePrefix, payload.StreamRoute, "delay")
	delayExchangeName := constructExchangeName(r.queuePrefix, payload.StreamRoute, "delay")

	zretry.DeclareExchange(channel, delayQueueName)
	zretry.QueueBind(channel, delayQueueName, delayExchangeName, amqp.Table{"x-dead-letter-exchange": constructExchangeName(r.queuePrefix, payload.StreamRoute, "dead_letter")})

	zretry.DeclareExchange(channel, exchangeName)
	_, queueDeclareErr := zretry.QueueDeclare(channel, queueName, nil)
	zlog.LogFatal(queueDeclareErr, "", nil)
	bindErr := zretry.QueueBind(channel, queueName, exchangeName, nil)
	zlog.LogFatal(bindErr, "", nil)
	if retryCountPayload >= r.retryCount {
		return zretry.PublishMessage(exchangeName, p, payload, expiration)
	}
	setRetryCount(&payload)
	return zretry.PublishMessage(exchangeName, p, payload, expiration)
}

var createPublisher = func(ctx context.Context, d *amqpextra.Dialer) (*publisher.Publisher, error) {
	options := []publisher.Option{publisher.WithContext(ctx)}
	return d.Publisher(options...)
}
