package rabbitmq

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestReplay(t *testing.T) {
	expectedCount := 10
	queueName := "bar"
	exchangeOut := "foo_exchange"
	expiry := "2"
	publishCallCount := 0
	channelGet = func(c *amqp.Channel, queueName string) (amqp.Delivery, bool, error) {
		return amqp.Delivery{
			Body: []byte{},
		}, true, nil
	}
	decodeMessage = func(body []byte) (zbasic.MessageEvent, error) {
		return zbasic.NewMessageEvent([]byte{}, []byte{}, "", "", "", time.Time{}), nil
	}
	publishMessage = func(exchangeName string, p *publisher.Publisher, payload zbasic.MessageEvent, expirationInMS string) error {
		publishCallCount++
		return nil
	}

	ackDelivery = func(d amqp.Delivery) error {
		return nil
	}

	p, _ := publisher.New(make(<-chan *publisher.Connection))
	err := replayMessages(&amqp.Channel{}, p, queueName, exchangeOut, expectedCount, expiry)
	if err != nil {
		t.Errorf("expected error to be nil but got %s", err.Error())
	}
	if publishCallCount != expectedCount {
		t.Errorf("expected call count to be %d but got %d", expectedCount, publishCallCount)
	}
}
