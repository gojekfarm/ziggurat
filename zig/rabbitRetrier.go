package zig

import (
	"bytes"
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
	connection     *amqp.Connection
	rabbitmqConfig *RabbitMQConfig
}

type RabbitMQConfig struct {
	host                 string
	delayQueueExpiration string
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
	retrierLogger.Info().Str("exchange-name", exchangeName).Msg("creating exchange")
	err := channel.ExchangeDeclare(exchangeName, amqp.ExchangeFanout, true, false, false, false, nil)
	return err
}

func createExchanges(channel *amqp.Channel, serviceName string, topicEntities []string, exchangeTypes []string) {
	for _, te := range topicEntities {
		for _, exchangeType := range exchangeTypes {
			exchangeName := constructExchangeName(serviceName, te, exchangeType)
			if err := createExchange(channel, exchangeName); err != nil {
				retrierLogger.Err(err).Msg("error creating exchange")
			}
		}
	}
}

func createAndBindQueue(channel *amqp.Channel, queueName string, exchangeName string, args amqp.Table) error {
	_, queueErr := channel.QueueDeclare(queueName, true, false, false, false, args)
	if queueErr != nil {
		return queueErr
	}
	retrierLogger.Info().Str("queue-name", queueName).Str("exchange-name", exchangeName).Msg("binding queue to exchange")
	bindErr := channel.QueueBind(queueName, "", exchangeName, false, nil)
	return bindErr
}

func createInstantQueues(channel *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, InstantType)
		exchangeName := constructExchangeName(serviceName, te, InstantType)
		if bindErr := createAndBindQueue(channel, queueName, exchangeName, nil); bindErr != nil {
			retrierLogger.Error().Err(bindErr).Msg("queue bind error")
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
			retrierLogger.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func createDeadLetterQueues(channel *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, DeadLetterType)
		exchangeName := constructExchangeName(serviceName, te, DeadLetterType)
		if bindErr := createAndBindQueue(channel, queueName, exchangeName, nil); bindErr != nil {
			retrierLogger.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func setRabbitMQConfig(config Config, r *RabbitRetrier) {
	rawConfig := config.GetByKey("rabbitmq")
	sanitizedConfig := rawConfig.(map[string]interface{})
	r.rabbitmqConfig = &RabbitMQConfig{
		host:                 sanitizedConfig["host"].(string),
		delayQueueExpiration: sanitizedConfig["delay-queue-expiration"].(string),
	}
}

func parseRabbitMQConfig(config *Config) *RabbitMQConfig {
	rawConfig := config.GetByKey("rabbitmq")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		retrierLogger.Error().Err(ErrParsingRabbitMQConfig).Msg("")
		return &RabbitMQConfig{
			host:                 "amqp://user:guest@localhost:5672/",
			delayQueueExpiration: "2000",
		}
	} else {
		return &RabbitMQConfig{
			host:                 sanitizedConfig["host"].(string),
			delayQueueExpiration: sanitizedConfig["delay-queue-expiration"].(string),
		}
	}
}

func (r *RabbitRetrier) Start(app *App) error {
	config := app.Config
	streamRoutes := app.StreamRouter.GetHandlerFunctionMap()
	rmqConfig := parseRabbitMQConfig(config)
	r.rabbitmqConfig = rmqConfig
	connection, err := amqp.Dial(r.rabbitmqConfig.host)
	if err != nil {
		return err
	}
	var topicEntities []string
	for te, _ := range streamRoutes {
		topicEntities = append(topicEntities, te)
	}
	r.connection = connection
	go func() {
		<-r.connection.NotifyClose(make(chan *amqp.Error))
		log.Error().Msg("rabbit retrier connection closed")
	}()
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
	if r.connection != nil {
		closeErr := r.connection.Close()
		return closeErr
	}
	return nil
}

func (r *RabbitRetrier) Retry(app *App, payload MessageEvent) error {
	config := app.Config
	channel, err := r.connection.Channel()
	exchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, DelayType)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, DeadLetterType)
	retryCount := getRetryCount(&payload)
	if retryCount == config.Retry.Count {
		err = publishMessage(channel, deadLetterExchangeName, payload, "")
		err = channel.Close()
		return err
	}
	setRetryCount(&payload)
	err = publishMessage(channel, exchangeName, payload, r.rabbitmqConfig.delayQueueExpiration)
	err = channel.Close()
	return err
}

func handleDelivery(app *App, ctag string, delivery <-chan amqp.Delivery, handlerFunc HandlerFunc, wg *sync.WaitGroup) {
	doneCh := app.Context().Done()
	for {
		select {
		case <-doneCh:
			retrierLogger.Info().Str("consumer-tag", ctag).Msg("stopping rabbit consumer")
			wg.Done()
			return
		case del := <-delivery:
			messageEvent, decodeErr := decodeMessage(del.Body)
			if decodeErr != nil {
				retrierLogger.Error().Err(decodeErr).Msg("retrier decode error")
			}
			if ackErr := del.Ack(false); ackErr != nil {
				retrierLogger.Error().Err(ackErr).Msg("rabbit retrier ack error")
			}
			retrierLogger.Info().Str("consumer-tag", ctag).Msg("handling rabbit message delivery")
			messageHandler(app, handlerFunc)(messageEvent)
		}
	}
}

func startRabbitConsumers(app *App, connection *amqp.Connection, config Config, topicEntity string, handlerFunc HandlerFunc, r *RabbitRetrier, wg *sync.WaitGroup) {
	channel, _ := connection.Channel()
	instantQueueName := constructQueueName(config.ServiceName, topicEntity, InstantType)
	ctag := topicEntity + "_amqp_consumer"
	deliveryChan, _ := channel.Consume(instantQueueName, ctag, false, false, false, false, nil)
	retrierLogger.Info().Str("consumer-tag", ctag).Msg("starting Rabbit consumer")
	go handleDelivery(app, ctag, deliveryChan, handlerFunc, wg)

}

func (r *RabbitRetrier) Consume(app *App) {
	streamRoutes := app.StreamRouter.GetHandlerFunctionMap()
	config := app.Config
	var wg sync.WaitGroup
	go func() {
		for teName, te := range streamRoutes {
			wg.Add(1)
			go startRabbitConsumers(app, r.connection, *config, teName, te.handlerFunc, r, &wg)
		}
		wg.Wait()
	}()
}

func decodeMessage(body []byte) (MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := &MessageEvent{Attributes: map[string]interface{}{}}
	if decodeErr := decoder.Decode(messageEvent); decodeErr != nil {
		return *messageEvent, decodeErr
	}
	return *messageEvent, nil
}

func handleReplayDelivery(r *RabbitRetrier, config Config, topicEntity string, deliveryChan <-chan amqp.Delivery, doneChan chan int) {
	channel, openErr := r.connection.Channel()
	if openErr != nil {
		retrierLogger.Error().Err(openErr)
		return
	}
	exchangeName := constructExchangeName(config.ServiceName, topicEntity, InstantType)
	defer channel.Close()
	for delivery := range deliveryChan {
		messageEvent, decodeErr := decodeMessage(delivery.Body)
		if decodeErr != nil {
			retrierLogger.Error().Err(decodeErr).Msg("rabbit retrier replay decode error")
		}
		publishErr := publishMessage(channel, exchangeName, messageEvent, r.rabbitmqConfig.delayQueueExpiration)
		if publishErr != nil {
			retrierLogger.Error().Err(publishErr).Msg("error publishing message")
		}
		if ackErr := delivery.Ack(false); ackErr != nil {
			retrierLogger.Error().Err(ackErr)
		}
	}
	close(doneChan)
}

func (r *RabbitRetrier) Replay(app *App, topicEntity string, count int) error {
	streamRoutes := app.StreamRouter.GetHandlerFunctionMap()
	config := app.Config
	if count == 0 {
		retrierLogger.Error().Err(ErrReplayCountZero).Msg("retrier replay error")
		return ErrReplayCountZero
	}
	if _, ok := streamRoutes[topicEntity]; !ok {
		retrierLogger.Error().Err(ErrTopicEntityMismatch).Msg("no topic entity found")
		return ErrTopicEntityMismatch
	}
	queueName := constructQueueName(config.ServiceName, topicEntity, DeadLetterType)
	channel, _ := r.connection.Channel()
	deliveryChan := make(chan amqp.Delivery, count)
	doneCh := make(chan int)
	go handleReplayDelivery(r, *config, topicEntity, deliveryChan, doneCh)
	for i := 0; i < count; i++ {
		delivery, _, _ := channel.Get(queueName, false)
		deliveryChan <- delivery
	}
	close(deliveryChan)
	<-doneCh
	return nil
}
