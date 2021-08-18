package rabbitmq

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"sync/atomic"
	"testing"
	"time"
)

func Test_RetryFlow(t *testing.T) {

	ctx, cfn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cfn()
	var callCount int32 = 0
	var expectedCallCount int32 = 5
	expectedValue := "foo"

	event := ziggurat.Event{
		Value: []byte(expectedValue),
	}
	ar := AutoRetry(
		WithQueues(QueueConfig{
			QueueName:           "test_12345",
			DelayExpirationInMS: "500",
			RetryCount:          5,
			WorkerCount:         1,
		}),
		WithUsername("user"),
		WithPassword("bitnami"),
		WithLogger(logger.NewDiscardLogger()))
	err := ar.InitPublishers(ctx)
	if err != nil {
		t.Errorf("could not init publishers %v", err)
	}
	done := make(chan struct{})
	go func() {
		err := ar.Stream(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
			if string(event.Value) == expectedValue {
				atomic.AddInt32(&callCount, 1)
			}
			return ziggurat.Retry
		}))
		if err != nil {
			t.Errorf("error running consumers: %v", err)
		}
		close(done)
	}()

	for i := 0; i < 5; i++ {
		err := ar.Publish(ctx, &event, "test_12345")
		if err != nil {
			t.Errorf("error publishing: %v", err)
		}
	}
	<-done

	if expectedCallCount != atomic.LoadInt32(&callCount) {
		t.Errorf("expected %d got %d", expectedCallCount, callCount)
	}

}
