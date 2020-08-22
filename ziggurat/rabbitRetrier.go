package ziggurat

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"sync"
)

const DelayType = "delay"
const InstantType = "instant"
const DeadLetterType = "dead_letter"
const RetryCount = "retryCount"

type RabbitRetrier struct {
	connection *amqp.Connection
}

func constructQueueName(serviceName string, topicEntity string, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", topicEntity, serviceName, queueType)
}

func constructExchangeName(serviceName string, topicEntity string, exchangeType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", topicEntity, serviceName, exchangeType)
}

func setRetryCount(m *MessageEvent) {
	value := m.GetMessageAttribute(RetryCount)
	if value == nil {
		m.SetMessageAttribute(RetryCount, 1)
		return
	}
	m.SetMessageAttribute(RetryCount, value.(int)+1)
}

func getRetryCount(m *MessageEvent) int {
	if value := m.GetMessageAttribute(RetryCount); value == nil {
		return 0
	}

	return m.GetMessageAttribute(RetryCount).(int)
}

func publishMessage(channel *amqp.Channel, exchangeName string, payload MessageEvent, expirationInMS string) error {
	buff := bytes.Buffer{}
	encoder := gob.NewEncoder(&buff)
	if encodeErr := encoder.Encode(payload); encodeErr != nil {
		return encodeErr
	}
	publishing := amqp.Publishing{
		Body:        buff.Bytes(),
		ContentType: "text/plain",
	}
	if expirationInMS != "" {
		publishing.Expiration = expirationInMS
	}
	if publishErr := channel.Publish(exchangeName, "", true, false, publishing); publishErr != nil {
		return publishErr
	}

	return nil
}

func createExchange(channel *amqp.Channel, exchangeName string) error {
	log.Info().Str("exchange-name", exchangeName).Msg("creating exchange")
	err := channel.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil)
	return err
}

func createExchanges(channel *amqp.Channel, serviceName string, topicEntities []string, exchangeTypes []string) {
	for _, te := range topicEntities {
		for _, exchangeType := range exchangeTypes {
			exchangeName := constructExchangeName(serviceName, te, exchangeType)
			if err := createExchange(channel, exchangeName); err != nil {
				log.Err(err).Msg("error creating exchange")
			}
		}
	}
}

func createAndBindQueue(channel *amqp.Channel, queueName string, exchangeName string, args amqp.Table) error {
	_, queueErr := channel.QueueDeclare(queueName, true, false, false, false, args)
	if queueErr != nil {
		return queueErr
	}
	log.Info().Str("queue-name", queueName).Str("exchange-name", exchangeName).Msg("binding queue to exchange")
	bindErr := channel.QueueBind(queueName, "", exchangeName, false, nil)
	return bindErr
}

func createInstantQueues(channel *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, InstantType)
		exchangeName := constructExchangeName(serviceName, te, InstantType)
		if bindErr := createAndBindQueue(channel, queueName, exchangeName, nil); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func createDelayQueues(channel *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, DelayType)
		exchangeName := constructExchangeName(serviceName, te, DelayType)
		deadLetterExchangeName := constructExchangeName(serviceName, te, InstantType)
		args := amqp.Table{
			"x-dead-letter-exchange": deadLetterExchangeName,
		}
		if bindErr := createAndBindQueue(channel, queueName, exchangeName, args); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func createDeadLetterQueues(channel *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, DeadLetterType)
		exchangeName := constructExchangeName(serviceName, te, DeadLetterType)
		if bindErr := createAndBindQueue(channel, queueName, exchangeName, nil); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func (r *RabbitRetrier) Start(config Config, streamRoutes TopicEntityHandlerMap) error {
	connection, err := amqp.Dial("amqp://user:bitnami@localhost:5672/")
	if err != nil {
		return err
	}
	var topicEntities []string
	for te, _ := range streamRoutes {
		topicEntities = append(topicEntities, te)
	}
	r.connection = connection
	channel, openErr := connection.Channel()
	if openErr != nil {
		return openErr
	}
	createExchanges(channel, config.ServiceName, topicEntities, []string{DelayType, InstantType, DeadLetterType})
	createInstantQueues(channel, topicEntities, config.ServiceName)
	createDelayQueues(channel, topicEntities, config.ServiceName)
	createDeadLetterQueues(channel, topicEntities, config.ServiceName)
	if closeErr := channel.Close(); closeErr != nil {
		return closeErr
	}
	return nil
}

func (r *RabbitRetrier) Stop() error {
	closeErr := r.connection.Close()
	return closeErr

}

func (r *RabbitRetrier) Retry(config Config, payload MessageEvent) error {
	channel, err := r.connection.Channel()
	exchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, DelayType)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, DeadLetterType)
	retryCount := getRetryCount(&payload)
	if retryCount == 5 {
		err = publishMessage(channel, deadLetterExchangeName, payload, "1000")
		err = channel.Close()
		return err
	}
	setRetryCount(&payload)
	err = publishMessage(channel, exchangeName, payload, "")
	err = channel.Close()
	return err
}

func handleDelivery(ctx context.Context, ctag string, delivery <-chan amqp.Delivery, config Config, r *RabbitRetrier, handlerFunc HandlerFunc, wg *sync.WaitGroup) {
	doneCh := ctx.Done()
	for {
		select {
		case <-doneCh:
			log.Info().Str("consumer-tag", ctag).Msg("stopping rabbit consumer")
			wg.Done()
			return
		case del := <-delivery:
			buff := bytes.Buffer{}
			buff.Write(del.Body)
			decoder := gob.NewDecoder(&buff)
			messageEvent := &MessageEvent{Attributes: map[string]interface{}{}}
			if decodeErr := decoder.Decode(messageEvent); decodeErr != nil {
				log.Error().Err(decodeErr).Msg("error decoding rabbitmq message payload")
				continue
			}
			log.Info().Str("consumer-tag", ctag).Msg("handling rabbit message delivery")
			MessageHandler(config, handlerFunc, r)(*messageEvent)
		}
	}
}

func startRabbitConsumers(ctx context.Context, connection *amqp.Connection, config Config, topicEntity string, handlerFunc HandlerFunc, r *RabbitRetrier, wg *sync.WaitGroup) {
	channel, _ := connection.Channel()
	instantQueueName := constructQueueName(config.ServiceName, topicEntity, InstantType)
	ctag := topicEntity + "_amqp_consumer"
	deliveryChan, _ := channel.Consume(instantQueueName, ctag, false, false, false, false, nil)
	log.Info().Str("consumer-tag", ctag).Msg("starting Rabbit consumer")
	wg.Add(1)
	go handleDelivery(ctx, ctag, deliveryChan, config, r, handlerFunc, wg)

}

func (r *RabbitRetrier) Consume(ctx context.Context, config Config, streamRoutes TopicEntityHandlerMap) {
	var wg sync.WaitGroup
	for teName, te := range streamRoutes {
		go startRabbitConsumers(ctx, r.connection, config, teName, te.handlerFunc, r, &wg)
	}
	wg.Wait()
}
