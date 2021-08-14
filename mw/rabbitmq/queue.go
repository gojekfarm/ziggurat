package rabbitmq

import "github.com/streadway/amqp"

func CreateAndBindQueue(ch *amqp.Channel, queueName string) error {
	if err := ch.ExchangeDeclare(queueName+"_exchange", amqp.ExchangeFanout, false, false, false, false, amqp.Table{}); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(queueName+"_queue", true, false, false, false, amqp.Table{}); err != nil {
		return err
	}

	if err := ch.QueueBind(queueName+"_queue", "", queueName+"_exchange", false, amqp.Table{}); err != nil {
		return err
	}
	return nil
}
