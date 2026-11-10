package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

func TestAutoRetry_KeepsEveryHost(t *testing.T) {
	r := AutoRetry(nil, WithHosts("a:5672", "b:5672"), WithUsername("u"), WithPassword("p"))

	want := []string{"amqp://u:p@a:5672", "amqp://u:p@b:5672"}
	if len(r.amqpURLs) != len(want) {
		t.Fatalf("expected %d urls, got %v", len(want), r.amqpURLs)
	}
	for i := range want {
		if r.amqpURLs[i] != want[i] {
			t.Fatalf("url %d: expected %q, got %q", i, want[i], r.amqpURLs[i])
		}
	}
}

func TestPublishInternal_PublishesPersistentMessages(t *testing.T) {
	for _, retryCount := range []int{5, 0} {
		p := &mockAMQPPublisher{}
		p.On("Publish", mock.MatchedBy(func(m publisher.Message) bool {
			return m.Publishing.DeliveryMode == amqp.Persistent
		})).Return(nil)

		if err := publishInternal(context.Background(), p, "foo", retryCount, "100", &ziggurat.Event{}); err != nil {
			t.Fatalf("retryCount %d: unexpected error %v", retryCount, err)
		}
		p.AssertExpectations(t)
	}
}

func TestRetry_UnreachableBrokerReturnsErrorAndRetriesInit(t *testing.T) {
	r := AutoRetry(Queues{{QueueKey: "foo", RetryCount: 1}},
		WithHosts("127.0.0.1:1"),
		WithConnectionTimeout(50*time.Millisecond),
	)

	for attempt := 1; attempt <= 2; attempt++ {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("attempt %d: Retry panicked: %v", attempt, rec)
				}
			}()
			if err := r.Retry(context.Background(), &ziggurat.Event{}, "foo"); err == nil {
				t.Fatalf("attempt %d: expected an error from an unreachable broker", attempt)
			}
		}()
		if r.publishDialer != nil {
			t.Fatalf("attempt %d: a failed init must leave the publishers uninitialised", attempt)
		}
	}
}
