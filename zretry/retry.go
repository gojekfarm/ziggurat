package zretry

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

type RabbitMQRetry struct {
	pdialer *amqpextra.Dialer
	cdialer *amqpextra.Dialer
	cfg     *RabbitMQConfig
}

func NewRabbitMQRetry(opts ...RabbitMQOptions) *RabbitMQRetry {
	rmq := &RabbitMQRetry{
		cfg: &RabbitMQConfig{},
	}
	for _, opt := range opts {
		opt(rmq.cfg)
	}

	if len(rmq.cfg.hosts) < 1 {
		rmq.cfg.hosts = "amqp://user:bitnami@localhost:5672/"
	}

	if rmq.cfg.queuePrefix == "" {
		rmq.cfg.queuePrefix = "rabbitmq"
	}
	if rmq.cfg.count == 0 {
		rmq.cfg.count = 1
	}
	if rmq.cfg.dialTimeoutInS == 0 {
		rmq.cfg.dialTimeoutInS = 30
	}
	if rmq.cfg.delayQueueExpiration == "" {
		rmq.cfg.delayQueueExpiration = "1000"
	}
	return rmq
}

func (R *RabbitMQRetry) Start(app ztype.App) error {
	routeNames := []string{}
	for _, route := range app.Routes() {
		routeNames = append(routeNames, route.RouteName)
	}
	var err error
	publishDialer, err := CreateDialer(app.Context(), splitHosts(R.cfg.hosts))
	if err != nil {
		return err
	}
	R.pdialer = publishDialer

	consumerDialer, err := CreateDialer(app.Context(), splitHosts(R.cfg.hosts))
	if err != nil {
		return err
	}
	R.cdialer = consumerDialer
	conn, err := GetConnectionFromDialer(app.Context(), publishDialer, time.Duration(R.cfg.dialTimeoutInS)*time.Second)
	if err != nil {
		return err
	}

	if err := withChannel(conn, func(c *amqp.Channel) error {
		createAndBindQueues(c, routeNames, R.cfg.queuePrefix)
		return nil
	}); err != nil {
		return err
	}

	if err := setupConsumers(app, consumerDialer, R.cfg.queuePrefix); err != nil {
		return err
	}
	return nil
}

func (R *RabbitMQRetry) Retry(app ztype.App, payload zbase.MessageEvent) error {
	ctx := app.Context()
	retryCount := R.cfg.count
	p, err := createPublisher(ctx, R.pdialer)
	if err != nil {
		return fmt.Errorf("error creating publisher: %s", err.Error())
	}
	defer p.Close()
	return retry(p, R.cfg.queuePrefix, retryCount, payload, R.cfg.delayQueueExpiration)
}

func (R *RabbitMQRetry) Stop() {
	if R.pdialer != nil {
		R.pdialer.Close()
	}

	if R.cdialer != nil {
		R.cdialer.Close()
	}
}

func (R *RabbitMQRetry) Replay(app ztype.App, route string, count int) error {
	p, perror := R.pdialer.Publisher(publisher.WithContext(app.Context()))
	if perror != nil {
		return perror
	}
	queueName := constructQueueName(R.cfg.queuePrefix, route, QueueTypeDL)
	exchangeOut := constructExchangeName(R.cfg.queuePrefix, route, QueueTypeInstant)
	conn, err := GetConnectionFromDialer(app.Context(), R.pdialer, 30*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	channelErr := withChannel(conn, func(c *amqp.Channel) error {
		return replayMessages(c, p, queueName, exchangeOut, count, R.cfg.delayQueueExpiration)
	})
	return channelErr
}
