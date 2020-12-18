package rabbitmq

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
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
	r := &RabbitMQRetry{
		hosts:       hosts,
		handler:     handler,
		queueConfig: queueConfig,
	}
	return r
}

func (r *RabbitMQRetry) HandleMessage(event ziggurat.MessageEvent, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
	status := r.handler.HandleMessage(event, z)
	if status == ziggurat.RetryMessage {
		ziggurat.LogFatal(r.retry(event, z), "rabbitmq failed to retry", nil)
	}
	return status
}

func (r *RabbitMQRetry) Retrier(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
		status := handler.HandleMessage(messageEvent, z)
		if status == ziggurat.RetryMessage {
			ziggurat.LogFatal(r.retry(messageEvent, z), "rabbitmq failed to retry", nil)
		}
		return status
	})
}

func (r *RabbitMQRetry) StartConsumers(z *ziggurat.Ziggurat, handler ziggurat.MessageHandler) error {
	consumerDialer, dialErr := amqpextra.NewDialer(
		amqpextra.WithContext(z.Context()),
		amqpextra.WithURL(r.hosts...))
	if dialErr != nil {
		return dialErr
	}
	r.consumerDialer = consumerDialer
	for routeName, _ := range r.queueConfig {
		queueName := constructQueueName(routeName, "instant")
		ctag := fmt.Sprintf("%s_%s_%s", queueName, "ziggurat", "ctag")
		c, err := createConsumer(z, r.consumerDialer, ctag, queueName, handler)
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
