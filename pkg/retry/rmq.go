package retry

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"strings"
	"time"
)

type RabbitMQConfig struct {
	Hosts                string `mapstructure:"hosts"`
	DelayQueueExpiration string `mapstructure:"delay-queue-expiration"`
	DialTimeoutInS       int    `mapstructure:"dial-timeout-seconds"`
}

func createContextWithDeadline(parentCtx context.Context, afterTimeInS int) (context.Context, context.CancelFunc) {
	deadlineTime := time.Now().Add(time.Duration(afterTimeInS) * time.Second)
	return context.WithDeadline(parentCtx, deadlineTime)
}

func parseRabbitMQConfig(config z.ConfigReader) *RabbitMQConfig {
	rmqcfg := &RabbitMQConfig{}
	if err := config.UnmarshalByKey("rabbitmq", rmqcfg); err != nil {
		logger.LogError(err, "rmq config unmarshall error", nil)
		return &RabbitMQConfig{
			Hosts:                "amqp://user:bitnami@localhost:5672/",
			DelayQueueExpiration: "2000",
		}
	}
	return rmqcfg
}

func splitHosts(hosts string) []string {
	return strings.Split(hosts, ",")
}

type RabbitMQRetry struct {
	pdialer *amqpextra.Dialer
	cdialer *amqpextra.Dialer
	cfg     *RabbitMQConfig
}

func NewRabbitMQRetry(config z.ConfigReader) *RabbitMQRetry {
	cfg := parseRabbitMQConfig(config)
	return &RabbitMQRetry{
		cfg: cfg,
	}
}

func withChannel(connection *amqp.Connection, cb func(c *amqp.Channel) error) error {
	c, err := connection.Channel()
	defer c.Close()
	if err != nil {
		return err
	}
	cberr := cb(c)
	return cberr
}

func createDialer(ctx context.Context, dialTimeoutInS int, hosts []string) (*amqpextra.Dialer, error) {
	d, cfgErr := amqpextra.NewDialer(
		amqpextra.WithURL(hosts...),
		amqpextra.WithContext(ctx))
	if cfgErr != nil {
		return nil, cfgErr
	}
	return d, nil
}

func (R *RabbitMQRetry) Start(app z.App) error {
	ctxWithDeadline, cancelFunc := createContextWithDeadline(app.Context(), R.cfg.DialTimeoutInS)
	defer cancelFunc()
	publishDialer, err := createDialer(ctxWithDeadline, R.cfg.DialTimeoutInS, splitHosts(R.cfg.Hosts))
	if err != nil {
		return err
	}
	R.pdialer = publishDialer

	consumerDialer, err := createDialer(ctxWithDeadline, R.cfg.DialTimeoutInS, splitHosts(R.cfg.Hosts))
	if err != nil {
		return err
	}
	R.cdialer = consumerDialer
	if err := setupConsumers(app, consumerDialer); err != nil {
		return err
	}

	conn, err := publishDialer.Connection(app.Context())
	if err != nil {
		return err
	}

	return withChannel(conn, func(c *amqp.Channel) error {
		createAndBindQueues(c, app.Router().GetTopicEntityNames(), app.Config().ServiceName)
		return nil
	})
}

func (R *RabbitMQRetry) Retry(app z.App, payload basic.MessageEvent) error {
	options := []publisher.Option{publisher.WithContext(app.Context())}
	p, err := R.pdialer.Publisher(options...)
	if err != nil {
		return err
	}
	defer p.Close()
	return retry(app.Context(), p, app.Config(), payload, R.cfg.DelayQueueExpiration)
}

func (R *RabbitMQRetry) Stop() error {
	if R.pdialer != nil {
		R.pdialer.Close()
	}

	if R.cdialer != nil {
		R.cdialer.Close()
	}
	return nil
}

func (R *RabbitMQRetry) Replay(app z.App, topicEntity string, count int) error {
	p, perror := R.pdialer.Publisher(publisher.WithContext(app.Context()))
	if perror != nil {
		return perror
	}
	queueName := constructQueueName(app.Config().ServiceName, topicEntity, QueueTypeDL)
	exchangeOut := constructExchangeName(app.Config().ServiceName, topicEntity, QueueTypeInstant)
	conn, err := R.pdialer.Connection(app.Context())
	if err != nil {
		return err
	}
	return withChannel(conn, func(c *amqp.Channel) error {
		return replayMessages(app.Context(), c, p, queueName, exchangeOut, count, R.cfg.DelayQueueExpiration)
	})
}
