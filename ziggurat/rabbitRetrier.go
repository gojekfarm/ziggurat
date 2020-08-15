package ziggurat

import (
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type RabbitRetrier struct {
	connection *amqp.Connection
}

func publishMessage(channel *amqp.Channel, exchangeName string, payload RetryPayload) error {
	publishing := amqp.Publishing{
		Body:        payload.MessageValueBytes,
		ContentType: "text/plain",
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

func createExchanges(channel *amqp.Channel, topicEntities []string, exchangeTypes []string) {
	for _, te := range topicEntities {
		for _, exchangeType := range exchangeTypes {
			exchangeName := te + "_" + exchangeType + "_" + "exchange"
			if err := createExchange(channel, exchangeName); err != nil {
				log.Err(err).Msg("error creating exchange")
			}
		}
	}
}

func createAndBindQueue(channel *amqp.Channel, queueName string, exchangeName string, queueType string) error {
	var args amqp.Table
	if queueType == "dead_letter" {
		args = amqp.Table{
			"x-dead-letter-exchange": exchangeName,
		}
	} else {
		args = nil
	}
	_, queueErr := channel.QueueDeclare(queueName, true, false, false, false, args)
	if queueErr != nil {
		return queueErr
	}
	bindErr := channel.QueueBind(queueName, "", exchangeName, false, nil)
	return bindErr
}

func createAndBindQueues(channel *amqp.Channel, topicEntities []string, queueTypes []string) {
	for _, te := range topicEntities {
		for _, qt := range queueTypes {
			queueName := te + "_" + qt + "_" + "queue"
			exchangeName := te + "_" + qt + "_" + "exchange"
			log.Info().Str("queue-name", queueName).Str("exchange-name", exchangeName).Msg("binding queue to exchange")
			if err := createAndBindQueue(channel, queueName, exchangeName, qt); err != nil {
				log.Error().Err(err).Msg("error creating queues")
			}
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
	createExchanges(channel, topicEntities, []string{"instant", "delay", "dead_letter"})
	createAndBindQueues(channel, topicEntities, []string{"instant", "delay", "dead_letter"})
	if closeErr := channel.Close(); closeErr != nil {
		return closeErr
	}
	return nil
}

func (r *RabbitRetrier) Stop() error {
	closeErr := r.connection.Close()
	return closeErr
}

func (r *RabbitRetrier) Retry(payload RetryPayload) error {
	channel, err := r.connection.Channel()
	err = publishMessage(channel, "test_exchange", payload)
	err = channel.Close()
	return err
}

func (r *RabbitRetrier) Consume(config Config, streamRoutes TopicEntityHandlerMap) {

}
