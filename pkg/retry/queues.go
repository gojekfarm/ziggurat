package retry

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/pkg/zlog"
	"github.com/streadway/amqp"
)

const QueueTypeDelay = "delay"
const QueueTypeInstant = "instant"
const QueueTypeDL = "dead_letter"

func constructQueueName(serviceName string, topicEntity string, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", topicEntity, serviceName, queueType)
}

func constructExchangeName(serviceName string, topicEntity string, exchangeType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", topicEntity, serviceName, exchangeType)
}

func declareExchanges(c *amqp.Channel, topicEntities []string, serviceName string) {
	exchangeTypes := []string{QueueTypeInstant, QueueTypeDelay, QueueTypeDL}
	for _, te := range topicEntities {
		for _, et := range exchangeTypes {
			exchName := constructExchangeName(serviceName, te, et)
			zlog.LogInfo("rmq queues: creating exchange", map[string]interface{}{"exchange-name": exchName})
			declareExchange(c, exchName)
		}
	}
}

func createAndBindQueue(c *amqp.Channel, queueName string, exchangeName string, args amqp.Table) error {
	_, queueErr := queueDeclare(c, queueName, args)
	if queueErr != nil {
		return queueErr
	}
	zlog.LogInfo("rmq queues: binding queue to exchange", map[string]interface{}{
		"queue-name":    queueName,
		"exchange-name": exchangeName,
	})
	bindErr := queueBind(c, queueName, exchangeName, args)
	return bindErr
}

func createInstantQueues(c *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeInstant)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		bindErr := createAndBindQueue(c, queueName, exchangeName, nil)
		zlog.LogError(bindErr, "rmq queues: error binding queue", nil)

	}
}

func createDelayQueues(c *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDelay)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDelay)
		deadLetterExchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		args := amqp.Table{
			"x-dead-letter-exchange": deadLetterExchangeName,
		}
		bindErr := createAndBindQueue(c, queueName, exchangeName, args)
		zlog.LogError(bindErr, "rmq queues: error binding queue", nil)
	}
}

func createDeadLetterQueues(c *amqp.Channel, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDL)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDL)
		bindErr := createAndBindQueue(c, queueName, exchangeName, nil)
		zlog.LogError(bindErr, "rmq queues: error binding queue", nil)
	}
}

var createAndBindQueues = func(c *amqp.Channel, topicEntities []string, serviceName string) {
	declareExchanges(c, topicEntities, serviceName)
	createInstantQueues(c, topicEntities, serviceName)
	createDelayQueues(c, topicEntities, serviceName)
	createDeadLetterQueues(c, topicEntities, serviceName)
}
