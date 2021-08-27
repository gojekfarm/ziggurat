package rabbitmq

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

func newAutoRetry(qn string) *autoRetry {
	ar := AutoRetry(
		[]QueueConfig{{
			QueueName:           qn,
			DelayExpirationInMS: "500",
			RetryCount:          5,
			WorkerCount:         1,
		}},
		WithUsername("user"),
		WithPassword("bitnami"),
		WithLogger(logger.NewDiscardLogger()))
	return ar
}

func Test_RetryFlow(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), 20*time.Second)
	defer cfn()
	var callCount int32 = 0
	var expectedCallCount int32 = 5
	expectedValue := "foo"
	queueName := "foo"

	event := ziggurat.Event{
		Value: []byte(expectedValue),
	}
	ar := newAutoRetry(queueName)
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

	for i := 0; i < int(expectedCallCount); i++ {
		err := ar.publish(ctx, &event, queueName)
		if err != nil {
			t.Errorf("error publishing: %v", err)
		}
	}
	<-done

	if expectedCallCount != atomic.LoadInt32(&callCount) {
		t.Errorf("expected %d got %d", expectedCallCount, callCount)
	}
}

func Test_view(t *testing.T) {
	qname := "blah"

	count := 5
	ctx, cfn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfn()
	type test struct {
		qname             string
		publishCount      int
		viewCount         int
		expectedViewCount int
		name              string
	}

	cases := []test{{
		name:              "read exact number of messages as there are in the queue",
		qname:             "foo_test",
		publishCount:      5,
		viewCount:         5,
		expectedViewCount: 5,
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ar := newAutoRetry(qname)
			for i := 0; i < c.publishCount; i++ {
				e := &ziggurat.Event{
					Value: []byte(fmt.Sprintf("bar-%d", i)),
				}
				err := ar.Publish(ctx, e, qname, "dlq", "")
				if err != nil {
					t.Errorf("error publishing to queue: %v", err)
				}
			}
			events, err := ar.view(ctx, qname, count, false)
			if err != nil {
				t.Errorf("error viewing messages: %v", err)
			}
			if len(events) != count {
				t.Errorf("expected to read %d messages but read %d", count, len(events))
			}
		})
	}

}
