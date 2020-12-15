package zmw

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/zrouter"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

type StreamRouteName string
type QueueConfig map[StreamRouteName]*struct {
	RetryCount               int
	DelayQueueExpirationInMS string
}

type RabbitMQRetry struct {
	hosts                []string
	dialer               *amqpextra.Dialer
	handler              ztype.MessageHandler
	dialTimeoutInSeconds time.Duration
	queueConfig          QueueConfig
}

func RabbitMQMW(ctx context.Context, hosts []string, queueConfig QueueConfig) zrouter.Adapter {
	r := &RabbitMQRetry{hosts: hosts, queueConfig: queueConfig}
	if err := r.connect(ctx); err != nil {
		zlog.LogFatal(err, "RABBITMQ CONNECTION ERROR", nil)
	}
	return func(next ztype.MessageHandler) ztype.MessageHandler {
		r.handler = next
		return r
	}
}

func (r *RabbitMQRetry) HandleMessage(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
	status := r.handler.HandleMessage(event, app)
	if status == ztype.RetryMessage {
		zlog.LogFatal(r.retry(event, app), "RABBITMQ RETRY MW", nil)
	}
	return status
}

func (r *RabbitMQRetry) connect(ctx context.Context) error {
	dialer, cfgErr := amqpextra.NewDialer(amqpextra.WithContext(ctx), amqpextra.WithURL(r.hosts...))
	if cfgErr != nil {
		return cfgErr
	}
	r.dialer = dialer
	conn, connErr := r.dialer.Connection(ctx)
	defer conn.Close()
	if connErr != nil {
		return connErr
	}

	channel, chanErr := conn.Channel()
	defer channel.Close()
	if chanErr != nil {
		return chanErr
	}

	queueTypes := []string{"instant", "delay", "dead_letter"}
	for _, queueType := range queueTypes {
		for route, _ := range r.queueConfig {
			exchangeName := constructExchangeName(route, queueType)
			if exchangeDeclareErr := channel.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil); exchangeDeclareErr != nil {
				return exchangeDeclareErr
			}

			queueName := constructQueueName(route, queueType)
			if _, queueDeclareErr := channel.QueueDeclare(queueName, true, false, false, false, nil); queueDeclareErr != nil {
				return queueDeclareErr
			}

			var args amqp.Table
			if queueType == "delay" {
				args = amqp.Table{"x-dead-letter-exchange": constructExchangeName(route, "instant")}
			}
			if bindErr := channel.QueueBind(queueName, "", exchangeName, false, args); bindErr != nil {
				return bindErr
			}
		}
	}
	return nil
}

func (r *RabbitMQRetry) retry(event zbase.MessageEvent, app ztype.App) error {
	pub, pubCreateError := r.dialer.Publisher(publisher.WithContext(app.Context()))
	if pubCreateError != nil {
		return pubCreateError
	}
	defer pub.Close()
	buff := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buff)
	publishing := amqp.Publishing{}
	message := publisher.Message{}

	if err := encoder.Encode(event); err != nil {
		return err
	}
	if getRetryCount(&event) >= r.queueConfig[StreamRouteName(event.StreamRoute)].RetryCount {
		message.Exchange = constructExchangeName(StreamRouteName(event.StreamRoute), "dead_letter")
		publishing.Expiration = ""
	} else {
		message.Exchange = constructExchangeName(StreamRouteName(event.StreamRoute), "delay")
		publishing.Expiration = r.queueConfig[StreamRouteName(event.StreamRoute)].DelayQueueExpirationInMS
		setRetryCount(&event)
	}
	publishing.Body = buff.Bytes()
	message.Publishing = publishing
	return pub.Publish(message)
}

func constructQueueName(routeName StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", "ziggurat", routeName, queueType)
}

func constructExchangeName(route StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", "ziggurat", route, queueType)
}

func getRetryCount(m *zbase.MessageEvent) int {
	if value := m.GetMessageAttribute("retryCount"); value == nil {
		return 0
	}
	return m.GetMessageAttribute("retryCount").(int)
}

func setRetryCount(m *zbase.MessageEvent) {
	value := m.GetMessageAttribute("retryCount")

	if value == nil {
		m.SetMessageAttribute("retryCount", 1)
		return
	}
	m.SetMessageAttribute("retryCount", value.(int)+1)
}
