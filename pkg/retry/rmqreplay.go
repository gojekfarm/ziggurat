package retry

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

func replayMessages(c *amqp.Channel, p *publisher.Publisher, queueName string, exchangeOutName string, count int, expiry string) error {
	for i := 0; i < count; i++ {
		delivery, ok, deliveryError := c.Get(queueName, false)
		if deliveryError != nil || !ok {
			zlogger.LogError(deliveryError, "rmq replay error", nil)
			return deliveryError
		}
		msg, decodeErr := decodeMessage(delivery.Body)
		if decodeErr != nil {
			return decodeErr
		}
		publishErr := publishMessage(exchangeOutName, p, msg, expiry)
		if publishErr != nil {
			return publishErr
		}
		if ackErr := delivery.Ack(false); ackErr != nil {
			zlogger.LogError(ackErr, "rmq replay ack error", nil)
		}
	}
	p.Close()
	return nil
}
