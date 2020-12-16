package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"strings"
	"time"
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
	handler        ztype.MessageHandler
	queueConfig    QueueConfig
}

func RabbitRetrier(ctx context.Context, hosts []string, queueConfig QueueConfig, handler ztype.MessageHandler) *RabbitMQRetry {
	r := &RabbitMQRetry{hosts: hosts, queueConfig: queueConfig, handler: handler}
	zlog.LogInfo("[RABBITMQ DIALING HOSTS]", map[string]interface{}{"HOSTS": strings.Join(hosts, ",")})
	if err := r.initPublisher(ctx); err != nil {
		zlog.LogFatal(err, "[RABBITMQ MW CONNECTION ERROR]", nil)
	}
	return r
}

func (r *RabbitMQRetry) HandleMessage(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
	status := r.handler.HandleMessage(event, app)
	if status == ztype.RetryMessage {
		zlog.LogFatal(r.retry(event, app), "RABBITMQ RETRY MW", nil)
	}
	return status
}

func (r *RabbitMQRetry) Retry(handler ztype.MessageHandler) ztype.MessageHandler {
	return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		status := handler.HandleMessage(messageEvent, app)
		if status == ztype.RetryMessage {
			zlog.LogFatal(r.retry(messageEvent, app), "RABBITMQ RETRY MW", nil)
		}
		return status
	})
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

func (r *RabbitMQRetry) retry(event zbase.MessageEvent, app ztype.App) error {
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

var decodeMessage = func(body []byte) (zbase.MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := zbase.NewMessageEvent(nil, nil, "", "", "", time.Time{})
	if decodeErr := decoder.Decode(&messageEvent); decodeErr != nil {
		return messageEvent, decodeErr
	}
	return messageEvent, nil
}

var createConsumer = func(app ztype.App, d *amqpextra.Dialer, ctag string, queueName string, msgHandler ztype.MessageHandler) (*consumer.Consumer, error) {
	options := []consumer.Option{
		consumer.WithInitFunc(func(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
			channel, err := conn.(*amqp.Connection).Channel()
			if err != nil {
				return nil, err
			}
			zlog.LogError(channel.Qos(1, 0, false), "rmq consumer: error setting QOS", nil)
			return channel, nil
		}),
		consumer.WithContext(app.Context()),
		consumer.WithConsumeArgs(ctag, false, false, false, false, nil),
		consumer.WithQueue(queueName),
		consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
			zlog.LogInfo("rmq consumer: processing message", map[string]interface{}{"QUEUE-NAME": queueName})
			msgEvent, err := decodeMessage(msg.Body)
			if err != nil {
				return msg.Reject(true)
			}
			msgHandler.HandleMessage(msgEvent, app)
			return msg.Ack(false)
		}))}
	return d.Consumer(options...)
}

func encodeMessage(message zbase.MessageEvent) (*bytes.Buffer, error) {
	buff := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(message); err != nil {
		return nil, err
	}
	return buff, nil
}

func (r *RabbitMQRetry) StartConsumers(app ztype.App, handler ztype.MessageHandler) error {
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
			zlog.LogError(fmt.Errorf("consumer closed"), "rmq consumer: closed", nil)
		}()
	}
	return nil
}
