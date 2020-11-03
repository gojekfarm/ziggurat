package zig

import (
	"fmt"
	amqpsafe "github.com/xssnick/amqp-safe"
)

type RabbitMQRetry struct {
	c      *amqpsafe.Connector
	config *RabbitMQConfig
}

func NewRabbitMQRetry(config *Config) *RabbitMQRetry {
	cfg := parseRabbitMQConfig(config)
	return &RabbitMQRetry{
		config: cfg,
	}
}

func (r *RabbitMQRetry) Start(app *App) (chan int, error) {
	conn := amqpsafe.NewConnector(amqpsafe.Config{
		Hosts: []string{r.config.host},
	})
	r.c = conn
	r.c.Start()
	setupCallback := createSetupCallback(r.c, app)
	r.c.OnReady(setupCallback)
	return make(chan int), nil
}

func (r *RabbitMQRetry) Retry(app *App, payload MessageEvent) error {
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

func (r *RabbitMQRetry) Replay(app *App, topicEntity string, count int) error {
	return nil
}
