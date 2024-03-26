package rabbitmq

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/streadway/amqp"
)

var queueTypes = []string{QueueTypeDelay, QueueTypeInstant, QueueTypeDL}

func createAndBindQueue(ch *amqp.Channel, queueName string, queueType string, args amqp.Table) error {
	queueWithType := fmt.Sprintf("%s_%s_%s", queueName, queueType, "queue")
	if _, err := ch.QueueDeclare(queueWithType, true, false, false, false, args); err != nil {
		return err
	}
	if err := ch.QueueBind(queueWithType, queueType, queueName+"_exchange", false, amqp.Table{}); err != nil {
		return err
	}
	return nil
}

func createQueuesAndExchanges(ch *amqp.Channel, queueName string, logger ziggurat.StructuredLogger) error {
	if err := ch.ExchangeDeclare(queueName+"_exchange", amqp.ExchangeDirect, true, false, false, false, amqp.Table{}); err != nil {
		return err
	}
	for _, qt := range queueTypes {
		args := amqp.Table{}
		logger.Info("creating queue", map[string]interface{}{"queue": queueName, "type": qt})
		if qt == QueueTypeDelay {
			args = amqp.Table{
				"x-dead-letter-exchange":    fmt.Sprintf("%s_%s", queueName, "exchange"),
				"x-dead-letter-routing-key": QueueTypeInstant,
			}
		}
		if err := createAndBindQueue(ch, queueName, qt, args); err != nil {
			return err
		}
	}
	return nil
}

func deleteQueuesAndExchanges(ch *amqp.Channel, queueName string) error {
	for _, qt := range queueTypes {
		exchangeName := fmt.Sprintf("%s_%s_%s", queueName, qt, "exchange")
		err := ch.ExchangeDelete(exchangeName, false, false)
		if err != nil {
			return err
		}
		queueName := fmt.Sprintf("%s_%s_%s", queueName, qt, "queue")
		_, err = ch.QueueDelete(queueName, false, false, false)
		if err != nil {
			return err
		}
	}
	return nil
}
