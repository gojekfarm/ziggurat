package zig

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

const DelayType = "delay"
const InstantType = "instant"
const DeadLetterType = "dead_letter"
const RetryCount = "retryCount"

type RabbitRetrier struct {
	pubConn          *amqp.Connection
	consumeConn      *amqp.Connection
	rabbitmqConfig   *RabbitMQConfig
	consumerStopChan chan int
}

type RabbitMQConfig struct {
	host                 string
	delayQueueExpiration string
}

func NewRabbitMQ(app *App) MessageRetrier {
	return &RabbitRetrier{
		rabbitmqConfig:   parseRabbitMQConfig(app.config),
		consumerStopChan: make(chan int),
	}
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

func publishConnectionCloseHandler(closeChan chan *amqp.Error) {
	err := <-closeChan
	retrierLogger.Error().Err(err).Msg("")
}

func consumeConnectionCloseHandler(closeChan chan *amqp.Error) {
	err := <-closeChan
	retrierLogger.Error().Err(err).Msg("")
}

func (r *RabbitRetrier) Start(app *App) (chan int, error) {
	config := app.config
	streamRoutes := app.router.GetHandlerFunctionMap()
	pubConn, err := amqp.Dial(r.rabbitmqConfig.host)
	consumeConn, err := amqp.Dial(r.rabbitmqConfig.host)
	r.pubConn = pubConn
	r.consumeConn = consumeConn
	if err != nil {
		return nil, err
	}
	var topicEntities []string
	for te, _ := range streamRoutes {
		topicEntities = append(topicEntities, te)
	}

	pubCloseChan := r.pubConn.NotifyClose(make(chan *amqp.Error))
	consumeCloseChan := r.consumeConn.NotifyClose(make(chan *amqp.Error))

	go publishConnectionCloseHandler(pubCloseChan)
	go consumeConnectionCloseHandler(consumeCloseChan)

	channel, openErr := pubConn.Channel()
	if openErr != nil {
		return nil, openErr
	}
	createExchanges(channel, config.ServiceName, topicEntities, []string{DelayType, InstantType, DeadLetterType})
	createInstantQueues(channel, topicEntities, config.ServiceName)
	createDelayQueues(channel, topicEntities, config.ServiceName)
	createDeadLetterQueues(channel, topicEntities, config.ServiceName)
	if closeErr := channel.Close(); closeErr != nil {
		return nil, closeErr
	}
	return r.StartConsumers(app), nil
}

func (r *RabbitRetrier) Stop() error {
	var closeErr error
	if r.pubConn != nil {
		closeErr = r.pubConn.Close()
	}
	if r.consumeConn != nil {
		closeErr = r.consumeConn.Close()
		return closeErr
	}
	return nil
}

func (r *RabbitRetrier) Retry(app *App, payload MessageEvent) error {
	config := app.config
	if !config.Retry.Enabled {
		retrierLogger.Fatal().Err(ErrRetryDisabled).Msg("[RETRIER ERROR]")
	}
	channel, err := r.pubConn.Channel()
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

func startRabbitConsumers(app *App, connection *amqp.Connection, config Config, topicEntity string, handlerFunc HandlerFunc, wg *sync.WaitGroup) {
	channel, _ := connection.Channel()
	instantQueueName := constructQueueName(config.ServiceName, topicEntity, InstantType)
	ctag := topicEntity + "_amqp_consumer"
	deliveryChan, _ := channel.Consume(instantQueueName, ctag, false, false, false, false, nil)
	retrierLogger.Info().Str("consumer-tag", ctag).Msg("starting Rabbit consumer")
	go handleDelivery(app, ctag, deliveryChan, handlerFunc, wg)

}

func (r *RabbitRetrier) StartConsumers(app *App) chan int {
	doneCh := make(chan int)
	streamRoutes := app.router.GetHandlerFunctionMap()
	config := app.config
	var wg sync.WaitGroup
	for teName, te := range streamRoutes {
		wg.Add(1)
		go startRabbitConsumers(app, r.consumeConn, *config, teName, te.handlerFunc, &wg)
	}
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	return doneCh
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

func handleReplayDelivery(ctx context.Context, r *RabbitRetrier, config Config, topicEntity string, deliveryChan <-chan amqp.Delivery, doneChan chan int) {
	channel, openErr := r.pubConn.Channel()
	if openErr != nil {
		retrierLogger.Error().Err(openErr)
		return
	}
	exchangeName := constructExchangeName(config.ServiceName, topicEntity, InstantType)
	defer channel.Close()
	ctxCancelChan := ctx.Done()
	for delivery := range deliveryChan {
		select {
		case <-ctxCancelChan:
			return
		default:
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
	}
	close(doneChan)
}

func (r *RabbitRetrier) Replay(app *App, topicEntity string, count int) error {
	streamRoutes := app.router.GetHandlerFunctionMap()
	config := app.config
	if count == 0 {
		retrierLogger.Error().Err(ErrReplayCountZero).Msg("retrier replay error")
		return ErrReplayCountZero
	}
	if _, ok := streamRoutes[topicEntity]; !ok {
		retrierLogger.Error().Err(ErrTopicEntityMismatch).Msg("no topic entity found")
		return ErrTopicEntityMismatch
	}
	queueName := constructQueueName(config.ServiceName, topicEntity, DeadLetterType)
	channel, _ := r.pubConn.Channel()
	deliveryChan := make(chan amqp.Delivery, count)
	doneCh := make(chan int)
	go handleReplayDelivery(app.Context(), r, *config, topicEntity, deliveryChan, doneCh)
	for i := 0; i < count; i++ {
		delivery, _, _ := channel.Get(queueName, false)
		deliveryChan <- delivery
	}
	close(deliveryChan)
	<-doneCh
	return nil
}
