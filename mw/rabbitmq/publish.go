package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

const HeaderRabbitMQ = "x-retry-agent"
const HeaderRabbitMQVal = "rabbitmq"

//mock the actual implementation
var publishAMQP = func(p *publisher.Publisher, msg publisher.Message) error {
	return p.Publish(msg)
}

func publish(p *publisher.Publisher, queue string, retryCount int, delayExpiration string, event *ziggurat.Event) error {

	expiration := delayExpiration

	if event.Metadata == nil {
		event.Metadata = map[string]interface{}{KeyRetryCount: 0}
	}

	newCount := getRetryCount(event) + 1

	exchange := fmt.Sprintf("%s_%s_%s", queue, "delay", "exchange")

	if newCount > retryCount {
		event.Metadata[KeyRetryCount] = retryCount
		exchange = fmt.Sprintf("%s_%s_%s", queue, "dlq", "exchange")
		expiration = ""
	} else {
		event.Metadata[KeyRetryCount] = newCount
	}

	event.Headers[HeaderRabbitMQ] = HeaderRabbitMQVal

	eb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := publisher.Message{
		Exchange: exchange,
		Publishing: amqp.Publishing{
			Expiration: expiration,
			Body:       eb,
		},
	}

	return publishAMQP(p, msg)
}
