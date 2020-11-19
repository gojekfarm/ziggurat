package retry

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"strings"
)

type RabbitMQConfig struct {
	Hosts                string `mapstructure:"hosts"`
	DelayQueueExpiration string `mapstructure:"delay-queue-expiration"`
	DialTimeoutInS       int    `mapstructure:"dial-timeout-seconds"`
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

func (R *RabbitMQRetry) Start(app z.App) error {
	var err error
	publishDialer, err := createDialer(app.Context(), splitHosts(R.cfg.Hosts))
	if err != nil {
		return err
	}
	R.pdialer = publishDialer

	consumerDialer, err := createDialer(app.Context(), splitHosts(R.cfg.Hosts))
	if err != nil {
		return err
	}
	R.cdialer = consumerDialer

	conn, err := getConnectionFromDialer(app.Context(), publishDialer)
	if err != nil {
		return err
	}

	if err := withChannel(conn, func(c *amqp.Channel) error {
		createAndBindQueues(c, app.Router().GetTopicEntityNames(), app.Config().ServiceName)
		return nil
	}); err != nil {
		return err
	}

	if err := setupConsumers(app, consumerDialer); err != nil {
		return err
	}
	return nil
}

func (R *RabbitMQRetry) Retry(app z.App, payload basic.MessageEvent) error {
	ctx := app.Context()
	p, err := createPublisher(ctx, R.pdialer)
	if err != nil {
		return fmt.Errorf("error creating publisher: %s", err.Error())
	}
	defer p.Close()
	return retry(p, app.Config(), payload, R.cfg.DelayQueueExpiration)
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
	conn, err := getConnectionFromDialer(app.Context(), R.pdialer)
	if err != nil {
		return err
	}
	return withChannel(conn, func(c *amqp.Channel) error {
		return replayMessages(app.Context(), c, p, queueName, exchangeOut, count, R.cfg.DelayQueueExpiration)
	})
}
