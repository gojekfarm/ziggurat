package zig

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	amqpsafe "github.com/xssnick/amqp-safe"
	"sync"
	"time"
)

var dialTimeout = 30 * time.Second

type RabbitMQRetry struct {
	pubConn  *amqpsafe.Connector
	config   *RabbitMQConfig
	consConn *amqpsafe.Connector
}

func NewRabbitMQRetry(config ConfigReader) *RabbitMQRetry {
	cfg := parseRabbitMQConfig(config)
	return &RabbitMQRetry{
		config: cfg,
	}
}

func (r *RabbitMQRetry) Start(app App) error {
	wg := &sync.WaitGroup{}
	connWaitChan := make(chan struct{})
	dialErrChan := make(chan error)
	r.pubConn = amqpsafe.NewConnector(amqpsafe.Config{
		Hosts: []string{r.config.host},
	})
	r.consConn = amqpsafe.NewConnector(amqpsafe.Config{
		Hosts: []string{r.config.host},
	})

	go func() {
		timerChan := time.After(dialTimeout)
		select {
		case <-timerChan:
			dialErrChan <- errors.New("dial timeout exceeded")
			r.pubConn.Close()
			r.consConn.Close()
		case <-connWaitChan:
			return
		}
	}()

	wg.Add(2)
	r.pubConn.Start().OnReady(func() {
		wg.Done()
	})
	setupCallback := createSetupCallback(r.consConn, app)
	r.consConn.Start().OnReady(func() {
		wg.Done()
		setupCallback()
	})

	go func() {
		wg.Wait()
		close(connWaitChan)
	}()

	done := app.Context().Done()
	select {
	case <-done:
		return nil
	case err := <-dialErrChan:
		return err
	case <-connWaitChan:
		logInfo("rmq: connected to rabbitmq!", map[string]interface{}{"host": r.config.host})
		return nil
	}

}

func (r *RabbitMQRetry) Retry(app App, payload MessageEvent) error {
	if app.Config().Retry.Enabled {
		return retry(app.Context(), r.pubConn, app.Config(), payload, r.config.delayQueueExpiration)
	}
	return fmt.Errorf("cannot retry message, `Retry.Enabled` is %v", app.Config().Retry.Enabled)
}

func (r *RabbitMQRetry) Stop() error {
	if r.pubConn != nil {
		return r.pubConn.Close()
	}
	if r.consConn != nil {
		return r.consConn.Close()
	}
	return nil
}

func (r *RabbitMQRetry) Replay(app App, topicEntity string, count int) error {
	// amqp-safe does not expose the `channel.Get` method,
	//dialing a new connection and using the `streadway/amqp` to consume single messages
	hfmap := app.Router().GetHandlerFunctionMap()
	if _, ok := hfmap[topicEntity]; !ok {
		return fmt.Errorf("error: topic-entity %s not registered", topicEntity)
	}
	if count < 1 {
		return fmt.Errorf("invalid count error: requested count %d is less than 1", count)
	}

	conn, dialErr := amqp.Dial(r.config.host)
	if dialErr != nil {
		return dialErr
	}

	channel, chanOpenErr := conn.Channel()
	if chanOpenErr != nil {
		return chanOpenErr
	}
	return replayMessages(app, r.pubConn, channel, topicEntity, count, r.config.delayQueueExpiration)
}
