package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

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

func createQueuesAndExchanges(ch *amqp.Channel, queueName string) error {
	queueTypes := []string{"delay", "instant", "dlq"}
	for _, qt := range queueTypes {
		args := amqp.Table{}
		qnameWithType := fmt.Sprintf("%s_%s", queueName, qt)
		if qt == "delay" {
			args = amqp.Table{"x-dead-letter-exchange": fmt.Sprintf("%s_%s_%s", queueName, "instant", "exchange")}
		}
		if err := createAndBindQueue(ch, qnameWithType, args); err != nil {
			return err
		}
	}
	return nil
}
