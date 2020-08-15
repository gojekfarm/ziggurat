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
	createExchanges(channel, topicEntities, []string{"instant", "delay", "dead_letter"})
	if openErr != nil {
		return openErr
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
	err = channel.Close()
	return err
}

func (r *RabbitRetrier) Consume(config Config, streamRoutes TopicEntityHandlerMap) {

}
