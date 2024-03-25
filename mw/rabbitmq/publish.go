package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

func publishInternal(p amqpPublisher, queue string, retryCount int, delayExpiration string, event *ziggurat.Event) error {

	expiration := delayExpiration

	if event.Metadata == nil {
		event.Metadata = map[string]interface{}{KeyRetryCount: 0}
	}

	newCount := RetryCountFor(event) + 1
	exchange := fmt.Sprintf("%s_%s", queue, "exchange")
	routingKey := QueueTypeDelay

	if newCount > retryCount {
		event.Metadata[KeyRetryCount] = retryCount
		routingKey = QueueTypeDL
		expiration = ""
	} else {
		event.Metadata[KeyRetryCount] = newCount
	}

	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := publisher.Message{
		Exchange: exchange,
		Key:      routingKey,
		Publishing: amqp.Publishing{
			Expiration: expiration,
			Body:       eb,
			Headers:    map[string]interface{}{"retry-origin": "ziggurat-go"},
		},
	}

	return p.Publish(msg)
}
