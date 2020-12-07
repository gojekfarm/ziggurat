package retry

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zlog"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

var channelGet = func(c *amqp.Channel, queueName string) (amqp.Delivery, bool, error) {
	return c.Get(queueName, false)
}

var ackDelivery = func(d amqp.Delivery) error {
	return d.Ack(false)
}

func replayMessages(c *amqp.Channel, p *publisher.Publisher, queueName string, exchangeOutName string, count int, expiry string) error {
	for i := 0; i < count; i++ {
		delivery, ok, deliveryError := channelGet(c, queueName)
		if deliveryError != nil || !ok {
			zlog.LogError(deliveryError, "rmq replay error", nil)
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
		if ackErr := ackDelivery(delivery); ackErr != nil {
			zlog.LogError(ackErr, "rmq replay ack error", nil)
		}
	}
	p.Close()
	return nil
}
