package rabbitmq

import (
	"testing"

	"github.com/gojekfarm/ziggurat"
	"github.com/makasim/amqpextra/publisher"
)

func Test_publish(t *testing.T) {
	var old = publishAMQP
	publishAMQP = func(p *publisher.Publisher, msg publisher.Message) error {
		return nil
	}
	defer func() {
		publishAMQP = old
	}()

	type test struct {
		name                 string
		input                *ziggurat.Event
		expectedExchangeName string
		retryCount           int
	}

	cases := []test{{
		name:                 "ARetry count",
		input:                &ziggurat.Event{},
		expectedExchangeName: "foo_exchange",
		retryCount:           1,
	}, {
		name:                 "ARetry count exceeded",
		input:                &ziggurat.Event{},
		expectedExchangeName: "foo_exchange",
		retryCount:           0,
	}}

	for _, c := range cases {
		p := publisher.Publisher{}
		publishAMQP = func(p *publisher.Publisher, msg publisher.Message) error {
			if msg.Exchange != c.expectedExchangeName {
				t.Errorf("expected exchange name %s got %s", c.expectedExchangeName, msg.Exchange)
			}
			return nil
		}
		_ = publishInternal(&p, "foo", c.retryCount, "100", c.input)
	}
}

func Test_publish_retries(t *testing.T) {
	var old = publishAMQP
	publishAMQP = func(p *publisher.Publisher, msg publisher.Message) error {
		return nil
	}
	p := publisher.Publisher{}
	defer func() {
		publishAMQP = old
	}()

	type test struct {
		name              string
		retryCount        int
		testFunc          func(t *testing.T, c test)
		expectedCount     int
		event             ziggurat.Event
		publishIterations int
	}
	cases := []test{{
		name:              "ARetry count increase",
		retryCount:        1,
		expectedCount:     1,
		publishIterations: 1,
	}, {
		name:              "ARetry count shouldn't exceed the max count",
		retryCount:        5,
		expectedCount:     5,
		event:             ziggurat.Event{},
		publishIterations: 7,
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := ziggurat.Event{}
			for i := 0; i < c.publishIterations; i++ {
				_ = publishInternal(&p, "", c.retryCount, "", &e)
			}
			if c.expectedCount != e.Metadata[KeyRetryCount] {
				t.Errorf("expected %v got %v", c.expectedCount, e.Metadata[KeyRetryCount])
			}
		})
	}
}
