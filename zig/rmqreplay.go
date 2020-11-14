package zig

import (
	"github.com/streadway/amqp"
)

func replayMessages(app App, channel *amqp.Channel, topicEntity string, count int, expiration string) error {
	//serviceName, ctx := app.Config().ServiceName, app.Context()
	//queueName := constructQueueName(serviceName, topicEntity, QueueTypeDL)
	//for i := 0; i < count; i++ {
	//	delivery, ok, deliveryErr := channel.Get(queueName, false)
	//	if !ok || deliveryErr != nil {
	//		return deliveryErr
	//	}
	//	exchangeName := constructExchangeName(serviceName, topicEntity, QueueTypeInstant)
	//	decodedMsg, decodeErr := decodeMessage(delivery.Body)
	//	if decodeErr != nil {
	//		logError(decodeErr, "ziggurat rmq replay", nil)
	//		return decodeErr
	//	}
	//	pubErr := publishMessage(ctx, connector, exchangeName, decodedMsg, expiration)
	//	if pubErr != nil {
	//		logError(pubErr, "ziggurat rmq replay", nil)
	//		return pubErr
	//	}
	//	if ackErr := delivery.Ack(false); ackErr != nil {
	//		logError(ackErr, "ziggurat rmq replay", nil)
	//		return ackErr
	//	}
	//}
	return nil
}
