package retry

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra"
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
	handler        ziggurat.Handler
	queueConfig    QueueConfig
	logger         ziggurat.LeveledLogger
}

func NewRabbitRetrier(hosts []string, queueConfig QueueConfig, logger ziggurat.LeveledLogger) *RabbitMQRetry {
	r := &RabbitMQRetry{
		hosts:       hosts,
		queueConfig: queueConfig,
		logger:      logger,
	}
	if r.logger == nil {
		r.logger = ziggurat.NewLogger("info")
	}
	return r
}

func (r *RabbitMQRetry) HandleMessage(event *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
	status := r.handler.HandleMessage(event, ctx)
	if status == ziggurat.RetryMessage {
		err := r.retry(event, ctx)
		r.logger.Errorf("error retrying message: %s", err)
	}
	return status
}

func (r *RabbitMQRetry) Retrier(handler ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		if r.dialer == nil {
			panic("dialer nil error: please start the call the `RunPublisher` method")
		}
		status := handler.HandleMessage(messageEvent, ctx)
		if status == ziggurat.RetryMessage {
			err := r.retry(messageEvent, ctx)
			r.logger.Errorf("error retrying message: %s", err)
		}
		return status
	})
}

func (r *RabbitMQRetry) RunPublisher(ctx context.Context) error {
	return r.initPublisher(ctx)
}

func (r *RabbitMQRetry) RunConsumers(ctx context.Context, handler ziggurat.Handler) error {
	consumerDialer, dialErr := amqpextra.NewDialer(
		amqpextra.WithContext(ctx),
		amqpextra.WithURL(r.hosts...))
	if dialErr != nil {
		return dialErr
	}
	r.consumerDialer = consumerDialer
	for routeName, _ := range r.queueConfig {
		queueName := constructQueueName(routeName, "instant")
		ctag := fmt.Sprintf("%s_%s_%s", queueName, "ziggurat", "ctag")
		c, err := createConsumer(ctx, r.consumerDialer, ctag, queueName, handler, r.logger)
		if err != nil {
			return err
		}
		go func() {
			<-c.NotifyClosed()
			r.logger.Error("consumer closed")
		}()
	}
	return nil
}

func (r *RabbitMQRetry) retry(event *ziggurat.Message, ctx context.Context) error {
	pub, pubCreateError := r.dialer.Publisher(publisher.WithContext(ctx))
	if pubCreateError != nil {
		return pubCreateError
	}
	defer pub.Close()

	publishing := amqp.Publishing{}
	message := publisher.Message{}

	if getRetryCount(event) >= r.queueConfig[StreamRouteName(event.RouteName)].RetryCount {
		message.Exchange = constructExchangeName(StreamRouteName(event.RouteName), "dead_letter")
		publishing.Expiration = ""
	} else {
		message.Exchange = constructExchangeName(StreamRouteName(event.RouteName), "delay")
		publishing.Expiration = r.queueConfig[StreamRouteName(event.RouteName)].DelayQueueExpirationInMS
		setRetryCount(event)
	}

	buff, err := encodeMessage(event)
	if err != nil {
		return err
	}

	publishing.Body = buff.Bytes()
	message.Publishing = publishing
	return pub.Publish(message)
}

func (r *RabbitMQRetry) initPublisher(ctx context.Context) error {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
	r.logger.Infof("dialing rabbitmq server with HOSTS=%s", strings.Join(r.hosts, ","))
	go func() {
		<-ctxWithTimeout.Done()
		cancelFunc()
	}()
	dialer, cfgErr := amqpextra.NewDialer(
		amqpextra.WithContext(ctx),
		amqpextra.WithURL(r.hosts...))
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
