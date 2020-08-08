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
		Body: payload.MessageValueBytes,
	}
	if publishErr := channel.Publish(exchangeName, "", true, false, publishing); publishErr != nil {
		return publishErr
	}

	return nil
}

func createExchange(channel *amqp.Channel, exchangeName string) error {
	log.Info().Str("exchange-name", exchangeName).Msg("creating exchange")
	err := channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	return err
}

func createAndBindQueue(channel *amqp.Channel, queueName string, exchangeName string) error {
	log.Info().Str("queue-name", queueName).Str("exchange-name", exchangeName).Msg("binding queue to exchange")
	_, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
	err = channel.QueueBind(queueName, "", exchangeName, false, nil)
	return err
}

func (r *RabbitRetrier) Start(config Config) error {
	connection, err := amqp.Dial("amqp://user:bitnami@localhost:5672/")
	if err != nil {
		return err
	}
	r.connection = connection
	channel, openErr := connection.Channel()
	if openErr != nil {
		return openErr
	}
	createErr := createExchange(channel, "test_exchange")
	if createErr != nil {
		return createErr
	}
	createAndBinErr := createAndBindQueue(channel, "test_queue", "test_exchange")
	if createAndBinErr != nil {
		return createAndBinErr
	}
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
	return err
}

func (r *RabbitRetrier) Consume(handlerFunc HandlerFunc) {

}
