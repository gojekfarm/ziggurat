package zig

import (
	"fmt"
	"github.com/streadway/amqp"
	amqpsafe "github.com/xssnick/amqp-safe"
)

type RabbitMQRetry struct {
	c      *amqpsafe.Connector
	config *RabbitMQConfig
}

func NewRabbitMQRetry(config ConfigReader) *RabbitMQRetry {
	cfg := parseRabbitMQConfig(config.Config())
	return &RabbitMQRetry{
		config: cfg,
	}
}

func (r *RabbitMQRetry) Start(app App) (chan int, error) {
	conn := amqpsafe.NewConnector(amqpsafe.Config{
		Hosts: []string{r.config.host},
	})
	r.c = conn
	r.c.Start()
	setupCallback := createSetupCallback(r.c, app)
	r.c.OnReady(setupCallback)
	return make(chan int), nil
}

func (r *RabbitMQRetry) Retry(app App, payload MessageEvent) error {
	if app.Config().Retry.Enabled {
		return retry(app.Context(), r.c, app.Config(), payload, r.config.delayQueueExpiration)
	}
	return fmt.Errorf("cannot retry message, `Retry.Enabled is %v", app.Config().Retry.Enabled)
}

func (r *RabbitMQRetry) Stop() error {
	if r.c != nil {
		return r.c.Close()
	}
	return nil
}

func (r *RabbitMQRetry) Replay(app App, topicEntity string, count int) error {
	// amqpsafe does not expose the `channel.Get` method,
	//dialing a new connection and using the `streadway/amqp` to consume single messages
	hfmap := app.Router().GetHandlerFunctionMap()
	if _, ok := hfmap[topicEntity]; !ok {
		return fmt.Errorf("error: topic-entity %s not registered", topicEntity)
	}
	if count < 1 {
		return fmt.Errorf("invalid count error: requested count %d is less than 1", count)
	}

	conn, err := amqp.Dial(r.config.host)
	if err != nil {
		return err
	}

	channel, chanOpenErr := conn.Channel()
	if chanOpenErr != nil {
		return chanOpenErr
	}
	return replayMessages(app, r.c, channel, topicEntity, count, r.config.delayQueueExpiration)
}
