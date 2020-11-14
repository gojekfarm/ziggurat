package zig

import (
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
)

type RMQExtra struct {
	dialer *amqpextra.Dialer
}

func NewRMQExtra(config ConfigReader) *RMQExtra {

	return &RMQExtra{}
}

func (R *RMQExtra) Start(app App) error {
	d, cfgErr := amqpextra.NewDialer(
		amqpextra.WithURL("amqp://user:bitnami@localhost:5672/"),
		amqpextra.WithContext(app.Context()))
	if cfgErr != nil {
		return cfgErr
	}
	R.dialer = d
	R.dialer.Notify()
	return nil
}

func (R *RMQExtra) Retry(app App, payload MessageEvent) error {
	p, err := R.dialer.Publisher(publisher.WithContext(app.Context()))
	logFatal(err, "", nil)
	return retry(app.Context(), p, app.Config(), payload, "1000")
}

func (R *RMQExtra) Stop() error {
	if R.dialer != nil {
		R.dialer.Close()
	}
	return nil
}

func (R *RMQExtra) Replay(app App, topicEntity string, count int) error {
	return nil
}
