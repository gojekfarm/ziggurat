package retry

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

func replayMessages(ctx context.Context, c *amqp.Channel, p *publisher.Publisher, queueName string, exchangeOutName string, count int, expiry string) error {
	for i := 0; i < count; i++ {
		delivery, ok, deliveryError := c.Get(queueName, false)
		if deliveryError != nil || !ok {
			logger.LogError(deliveryError, "rmq replay error", nil)
			return deliveryError
		}
		msg, decodeErr := decodeMessage(delivery.Body)
		if decodeErr != nil {
			return decodeErr
		}
		publishErr := publishMessage(ctx, exchangeOutName, p, msg, expiry)
		if publishErr != nil {
			return publishErr
		}
		if ackErr := delivery.Ack(false); ackErr != nil {
			logger.LogError(ackErr, "rmq replay ack error", nil)
		}
	}
	p.Close()
	return nil
}
