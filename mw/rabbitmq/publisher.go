package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"time"
)

func constructQueueName(routeName StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", "ziggurat", routeName, queueType)
}

func constructExchangeName(route StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", "ziggurat", route, queueType)
}

func getRetryCount(m *ziggurat.MessageEvent) int {
	if value := m.GetMessageAttribute("retryCount"); value == nil {
		return 0
	}
	return m.GetMessageAttribute("retryCount").(int)
}

func setRetryCount(m *ziggurat.MessageEvent) {
	value := m.GetMessageAttribute("retryCount")

	if value == nil {
		m.SetMessageAttribute("retryCount", 1)
		return
	}
	m.SetMessageAttribute("retryCount", value.(int)+1)
}

func encodeMessage(message ziggurat.MessageEvent) (*bytes.Buffer, error) {
	buff := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(message); err != nil {
		return nil, err
	}
	return buff, nil
}

func (r *RabbitMQRetry) initPublisher(ctx context.Context) error {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
	go func() {
		<-ctxWithTimeout.Done()
		cancelFunc()
	}()
	dialer, cfgErr := amqpextra.NewDialer(amqpextra.WithContext(ctx), amqpextra.WithURL(r.hosts...))
	if cfgErr != nil {
		return cfgErr
	}
	r.dialer = dialer

	conn, connErr := r.dialer.Connection(ctxWithTimeout)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()

	channel, chanErr := conn.Channel()
	if chanErr != nil {
		return chanErr
	}
	defer channel.Close()

	queueTypes := []string{"instant", "delay", "dead_letter"}
	for _, queueType := range queueTypes {
		var args amqp.Table
		for route, _ := range r.queueConfig {
			if queueType == "delay" {
				args = amqp.Table{"x-dead-letter-exchange": constructExchangeName(route, "instant")}
			}
			exchangeName := constructExchangeName(route, queueType)
			if exchangeDeclareErr := channel.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil); exchangeDeclareErr != nil {
				return exchangeDeclareErr
			}

			queueName := constructQueueName(route, queueType)
			if _, queueDeclareErr := channel.QueueDeclare(queueName, true, false, false, false, args); queueDeclareErr != nil {
				return queueDeclareErr
			}

			if bindErr := channel.QueueBind(queueName, "", exchangeName, false, args); bindErr != nil {
				return bindErr
			}
		}
	}
	return nil
}

func (r *RabbitMQRetry) retry(event ziggurat.MessageEvent, app ziggurat.App) error {
	pub, pubCreateError := r.dialer.Publisher(publisher.WithContext(app.Context()))
	if pubCreateError != nil {
		return pubCreateError
	}
	defer pub.Close()

	publishing := amqp.Publishing{}
	message := publisher.Message{}

	if getRetryCount(&event) >= r.queueConfig[StreamRouteName(event.StreamRoute)].RetryCount {
		message.Exchange = constructExchangeName(StreamRouteName(event.StreamRoute), "dead_letter")
		publishing.Expiration = ""
	} else {
		message.Exchange = constructExchangeName(StreamRouteName(event.StreamRoute), "delay")
		publishing.Expiration = r.queueConfig[StreamRouteName(event.StreamRoute)].DelayQueueExpirationInMS
		setRetryCount(&event)
	}

	buff, err := encodeMessage(event)
	if err != nil {
		return err
	}

	publishing.Body = buff.Bytes()
	message.Publishing = publishing
	return pub.Publish(message)
}
