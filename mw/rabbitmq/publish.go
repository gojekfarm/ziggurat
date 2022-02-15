package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

//mock the actual implementation
var publishAMQP = func(p *publisher.Publisher, msg publisher.Message) error {
	return p.Publish(msg)
}

func publishInternal(p *publisher.Publisher, queue string, retryCount int, delayExpiration string, event *ziggurat.Event) error {

	expiration := delayExpiration

	if event.Metadata == nil {
		event.Metadata = map[string]interface{}{KeyRetryCount: 0}
	}

	newCount := getRetryCount(event) + 1
	exchange := fmt.Sprintf("%s_%s", queue, "exchange")
	routingKey := QueueDelay

	if newCount > retryCount {
		event.Metadata[KeyRetryCount] = retryCount
		routingKey = QueueDL
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

	return publishAMQP(p, msg)
}
