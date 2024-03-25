package rabbitmq

import (
	"encoding/json"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"testing"
)

func Test_publish(t *testing.T) {

	type test struct {
		name                 string
		input                *ziggurat.Event
		expectedExchangeName string
		retryCount           int
		WantMsg              publisher.Message
	}

	cases := []test{{
		name:                 "ARetry count",
		input:                &ziggurat.Event{},
		expectedExchangeName: "foo_exchange",
		retryCount:           1,
		WantMsg: publisher.Message{
			Exchange: "foo_exchange",
			Key:      "delay",
			Publishing: amqp.Publishing{
				Headers:    map[string]interface{}{"retry-origin": "ziggurat-go"},
				Expiration: "100",
				Body: toJSON(ziggurat.Event{
					Metadata: map[string]any{KeyRetryCount: 1},
				}),
			},
		},
	}, {
		name:                 "ARetry count exceeded",
		input:                &ziggurat.Event{},
		expectedExchangeName: "foo_exchange",
		retryCount:           0,
		WantMsg: publisher.Message{
			Exchange: "foo_exchange",
			Key:      "dlq",
			Publishing: amqp.Publishing{
				Headers:    map[string]interface{}{"retry-origin": "ziggurat-go"},
				Expiration: "",
				Body:       toJSON(ziggurat.Event{Metadata: map[string]any{KeyRetryCount: 0}}),
			},
		},
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := mockAMQPPublisher{}
			p.On("Publish", c.WantMsg).Return(nil)
			_ = publishInternal(&p, "foo", c.retryCount, "100", c.input)
		})
	}
}

func Test_publish_retries(t *testing.T) {

	type test struct {
		name              string
		retryCount        int
		expectedCount     int
		event             ziggurat.Event
		publishIterations int
		WantMsg           publisher.Message
	}

	cases := []test{{
		name:              "ARetry count increase",
		retryCount:        1,
		expectedCount:     1,
		publishIterations: 1,
		WantMsg: publisher.Message{
			Exchange: "foo_exchange",
			Key:      QueueTypeDelay,
			Publishing: amqp.Publishing{
				Expiration: "100",
				Body: toJSON(ziggurat.Event{
					Metadata: map[string]any{KeyRetryCount: 1},
				}),
				Headers: map[string]interface{}{"retry-origin": "ziggurat-go"},
			},
		},
	}, {
		name:              "ARetry count shouldn't exceed the max count",
		retryCount:        5,
		expectedCount:     5,
		event:             ziggurat.Event{},
		publishIterations: 7,
		WantMsg: publisher.Message{
			Exchange: "foo_exchange",
			Key:      QueueTypeDelay,
			Publishing: amqp.Publishing{
				Expiration: "100",
				Headers:    map[string]interface{}{"retry-origin": "ziggurat-go"},
			},
		},
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := mockAMQPPublisher{}
			e := ziggurat.Event{}

			for i := 0; i < c.publishIterations; i++ {
				want := c.WantMsg
				retryCount := i + 1
				if retryCount > c.retryCount {
					want.Key = QueueTypeDL
					want.Publishing.Expiration = ""
					retryCount = c.retryCount
				}
				want.Publishing.Body = toJSON(ziggurat.Event{Metadata: map[string]any{KeyRetryCount: retryCount}})
				p.On("Publish", want).Return(nil).Once()
				_ = publishInternal(&p, "foo", c.retryCount, "100", &e)
			}
		})
	}
}

func toJSON(v any) []byte {
	bb, _ := json.Marshal(v)
	return bb
}
