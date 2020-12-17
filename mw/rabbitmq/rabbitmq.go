package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"strings"
)

type StreamRouteName string
type QueueConfig map[StreamRouteName]*struct {
	RetryCount               int
	DelayQueueExpirationInMS string
}

type RabbitMQRetry struct {
	hosts          []string
	dialer         *amqpextra.Dialer
	consumerDialer *amqpextra.Dialer
	handler        ziggurat.MessageHandler
	queueConfig    QueueConfig
}

func NewRabbitRetrier(ctx context.Context, hosts []string, queueConfig QueueConfig, handler ziggurat.MessageHandler) *RabbitMQRetry {
	r := &RabbitMQRetry{hosts: hosts, queueConfig: queueConfig, handler: handler}
	ziggurat.LogInfo("rabbitmq dialing hosts", map[string]interface{}{"HOSTS": strings.Join(hosts, ",")})
	if err := r.initPublisher(ctx); err != nil {
		ziggurat.LogFatal(err, "rabbitmq init error", nil)
	}
	return r
}

func (r *RabbitMQRetry) HandleMessage(event ziggurat.MessageEvent, app ziggurat.App) ziggurat.ProcessStatus {
	status := r.handler.HandleMessage(event, app)
	if status == ziggurat.RetryMessage {
		ziggurat.LogFatal(r.retry(event, app), "rabbitmq failed to retry", nil)
	}
	return status
}

func (r *RabbitMQRetry) Retrier(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, app ziggurat.App) ziggurat.ProcessStatus {
		status := handler.HandleMessage(messageEvent, app)
		if status == ziggurat.RetryMessage {
			ziggurat.LogFatal(r.retry(messageEvent, app), "rabbitmq failed to retry", nil)
		}
		return status
	})
}

func (r *RabbitMQRetry) StartConsumers(app ziggurat.App, handler ziggurat.MessageHandler) error {
	consumerDialer, dialErr := amqpextra.NewDialer(
		amqpextra.WithContext(app.Context()),
		amqpextra.WithURL(r.hosts...))
	if dialErr != nil {
		return dialErr
	}
	r.consumerDialer = consumerDialer
	for routeName, _ := range r.queueConfig {
		queueName := constructQueueName(routeName, "instant")
		ctag := fmt.Sprintf("%s_%s_%s", queueName, "ziggurat", "ctag")
		c, err := createConsumer(app, r.consumerDialer, ctag, queueName, handler)
		if err != nil {
			return err
		}
		go func() {
			<-c.NotifyClosed()
			ziggurat.LogError(fmt.Errorf("consumer closed"), "rmq consumer: closed", nil)
		}()
	}
	return nil
}
