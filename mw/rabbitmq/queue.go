package rabbitmq

import (
	"fmt"

	"github.com/gojekfarm/ziggurat"
	"github.com/streadway/amqp"
)

var queueTypes = []string{"delay", "instant", "dlq"}

func createAndBindQueue(ch *amqp.Channel, queueName string, args amqp.Table) error {
	if err := ch.ExchangeDeclare(queueName+"_exchange", amqp.ExchangeFanout, false, false, false, false, amqp.Table{}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueName+"_queue", true, false, false, false, args); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName+"_queue", "", queueName+"_exchange", false, amqp.Table{}); err != nil {
		return err
	}
	return nil
}

func createQueuesAndExchanges(ch *amqp.Channel, queueName string, logger ziggurat.StructuredLogger) error {
	for _, qt := range queueTypes {
		args := amqp.Table{}
		qnameWithType := fmt.Sprintf("%s_%s", queueName, qt)
		logger.Info("creating queue", map[string]interface{}{"queue": qnameWithType})
		if qt == "delay" {
			args = amqp.Table{"x-dead-letter-exchange": fmt.Sprintf("%s_%s_%s", queueName, "instant", "exchange")}
		}
		if err := createAndBindQueue(ch, qnameWithType, args); err != nil {
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
