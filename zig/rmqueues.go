package zig

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	amqpsafe "github.com/xssnick/amqp-safe"
)

const QueueTypeDelay = "delay"
const QueueTypeInstant = "instant"
const QueueTypeDL = "dead_letter"

func declareExchanges(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	exchangeTypes := []string{QueueTypeInstant, QueueTypeDelay, QueueTypeDL}
	for _, te := range topicEntities {
		for _, et := range exchangeTypes {
			exchName := constructExchangeName(serviceName, te, et)
			log.Info().Str("exchange-name", exchName).Msg("creating exchange")
			c.ExchangeDeclare(exchName, amqpsafe.ExchangeFanout, true, false, false, false, nil)
		}
	}

}

func createAndBindQueue(c *amqpsafe.Connector, queueName string, exchangeName string, args amqp.Table) error {
	_, queueErr := c.QueueDeclare(queueName, true, false, false, false, args)
	if queueErr != nil {
		return queueErr
	}
	log.Info().Str("queue-name", queueName).Str("exchange-name", exchangeName).Msg("binding queue to exchange")
	bindErr := c.QueueBind(queueName, "", exchangeName, false, nil)
	return bindErr
}

func constructQueueName(serviceName string, topicEntity string, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", topicEntity, serviceName, queueType)
}

func constructExchangeName(serviceName string, topicEntity string, exchangeType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", topicEntity, serviceName, exchangeType)
}

func createInstantQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeInstant)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		if bindErr := createAndBindQueue(c, queueName, exchangeName, nil); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func createDelayQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDelay)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDelay)
		deadLetterExchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		args := amqp.Table{
			"x-dead-letter-exchange": deadLetterExchangeName,
		}
		if bindErr := createAndBindQueue(c, queueName, exchangeName, args); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}

func createDeadLetterQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDL)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDL)
		if bindErr := createAndBindQueue(c, queueName, exchangeName, nil); bindErr != nil {
			log.Error().Err(bindErr).Msg("queue bind error")
		}
	}
}
