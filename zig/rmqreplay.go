package zig

import (
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	amqpsafe "github.com/xssnick/amqp-safe"
)

func replayMessages(app *App, connector *amqpsafe.Connector, channel *amqp.Channel, topicEntity string, count int, expiration string) error {
	serviceName, ctx := app.Config().ServiceName, app.Context()
	queueName := constructQueueName(serviceName, topicEntity, QueueTypeDL)
	for i := 0; i < count; i++ {
		delivery, ok, deliveryErr := channel.Get(queueName, false)
		if !ok || deliveryErr != nil {
			return deliveryErr
		}
		exchangeName := constructExchangeName(serviceName, topicEntity, QueueTypeInstant)
		decodedMsg, decodeErr := decodeMessage(delivery.Body)
		if decodeErr != nil {
			log.Error().Err(decodeErr).Msg("[RABBITMQ REPLAY]")
			return decodeErr
		}
		pubErr := publishMessage(ctx, connector, exchangeName, decodedMsg, expiration)
		if pubErr != nil {
			log.Error().Err(pubErr).Msg("[RABBITMQ REPLAY]")
			return pubErr
		}
		if ackErr := delivery.Ack(false); ackErr != nil {
			log.Error().Err(ackErr).Msg("[RABBITMQ REPLAY]")
			return ackErr
		}
	}
	return nil
}
