package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat"
)

func newAutoRetry(qn string, count int, consumerCount int) *autoRetry {
	ar := AutoRetry(
		[]QueueConfig{{
			QueueName:           qn,
			DelayExpirationInMS: "500",
			RetryCount:          count,
			ConsumerCount:       consumerCount,
		}},
		WithUsername("user"),
		WithConnectionTimeout(5*time.Second),
		WithPassword("bitnami"))
	return ar
}

func Test_RetryFlow(t *testing.T) {

	type test struct {
		PublishCount  int
		RetryCount    int
		QueueName     string
		Name          string
		ConsumerCount int
	}

	cases := []test{
		{
			PublishCount: 20,
			RetryCount:   5,
			Name:         "handler is called for PublishCount * RetryCount times",
			QueueName:    "foo",
		},
		{
			PublishCount: 10,
			RetryCount:   5,
			Name:         "spawns one consumer when the count is 0",
			QueueName:    "bar",
		},
		{
			PublishCount:  10,
			RetryCount:    2,
			Name:          "expect handler to be called PublishCount*RetryCount times with multiple consumers",
			QueueName:     "baz",
			ConsumerCount: 25,
		},
	}

	ctx, cfn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfn()

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var callCount int32
			expectedCallCount := int32(c.PublishCount * c.RetryCount)
			ar := newAutoRetry(c.QueueName, c.RetryCount, c.ConsumerCount)
			err := ar.InitPublishers(ctx)
			if err != nil {
				t.Errorf("could not init publishers %v", err)
			}
			done := make(chan struct{})
			go func() {
				err := ar.Stream(ctx, ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
					atomic.AddInt32(&callCount, 1)
					return ziggurat.Retry
				}, c.QueueName))
				if !errors.Is(err, ErrCleanShutdown) {
					t.Errorf("error running consumers: %v", err)
				}
				close(done)
			}()

			for i := 0; i < c.PublishCount; i++ {
				err := ar.publish(ctx, &ziggurat.Event{Value: []byte(fmt.Sprintf("foo-%d", i))}, c.QueueName)
				if err != nil {
					t.Errorf("error publishing: %v", err)
				}
			}
			<-done

			if expectedCallCount != atomic.LoadInt32(&callCount) {
				t.Errorf("expected %d got %d", expectedCallCount, callCount)
			}
			err = ar.DeleteQueuesAndExchanges(context.Background(), c.QueueName)
			if err != nil {
				t.Errorf("error deleting queues:%v", err)
			}
		})
	}

}

func Test_view(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfn()
	type test struct {
		qname             string
		publishCount      int
		viewCount         int
		expectedViewCount int
		name              string
	}

	cases := []test{
		{
			name:              "read exact number of messages as there are in the queue",
			qname:             "foo_test",
			publishCount:      5,
			viewCount:         5,
			expectedViewCount: 5,
		},
		{
			name:              "read excess number of messages than there are in the queue",
			qname:             "bar_test",
			publishCount:      5,
			viewCount:         10,
			expectedViewCount: 5,
		},
		{
			name:              "read negative number of messages",
			qname:             "baz_test",
			publishCount:      5,
			viewCount:         -1,
			expectedViewCount: 0,
		},
		{
			name:              "read zero messages",
			qname:             "foo_test",
			viewCount:         0,
			publishCount:      5,
			expectedViewCount: 0,
		}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ar := newAutoRetry(c.qname, 5, 1)
			err := ar.InitPublishers(ctx)
			if err != nil {
				t.Errorf("error could not init publishers:%v", err)
			}
			for i := 0; i < c.publishCount; i++ {
				e := &ziggurat.Event{
					Value: []byte(fmt.Sprintf("bar-%d", i)),
				}
				err := ar.Publish(ctx, e, c.qname, "dlq", "")
				if err != nil {
					t.Errorf("error publishing to queue: %v", err)
				}
			}
			events, err := ar.view(ctx, c.qname, c.viewCount, false)
			if err != nil {
				t.Errorf("error viewing messages: %v", err)
			}
			if len(events) != c.expectedViewCount {
				t.Errorf("expected to read %d messages but read %d", c.expectedViewCount, len(events))
			}
			err = ar.DeleteQueuesAndExchanges(context.Background(), c.qname)
			if err != nil {
				t.Errorf("error deleting queues:%v", err)
			}
		})
	}

}

func Test_replay(t *testing.T) {
	type test struct {
		publishCount  int
		replayCount   int
		expectedCount int
		queueName     string
		name          string
	}

	cases := []test{
		{
			name:          "replay exact number of messages",
			publishCount:  5,
			replayCount:   5,
			expectedCount: 5,
			queueName:     "foo_test",
		},
		{
			name:          "replay excess number of messages",
			publishCount:  5,
			replayCount:   10,
			expectedCount: 5,
			queueName:     "foo_test",
		},
		{
			name:          "replay 0 messages",
			publishCount:  5,
			replayCount:   0,
			expectedCount: 0,
			queueName:     "foo_test",
		},
		{
			name:          "replay negative number of messages",
			publishCount:  5,
			replayCount:   -1,
			expectedCount: 0,
			queueName:     "foo_test",
		}}

	ctx, cfn := context.WithTimeout(context.Background(), time.Second*10)

	defer cfn()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ar := newAutoRetry(tc.queueName, 5, 1)
			err := ar.InitPublishers(ctx)
			if err != nil {
				t.Errorf("couldn't start publishers:%v", err)
			}
			for i := 0; i < tc.publishCount; i++ {
				e := ziggurat.Event{Value: []byte(fmt.Sprintf("%s-%d", "foo", i))}
				err := ar.Publish(ctx, &e, tc.queueName, "dlq", "")
				if err != nil {
					t.Errorf("error publishing to dql:%v", err)
				}
			}

			c, err := ar.replay(ctx, tc.queueName, tc.replayCount)
			if err != nil {
				t.Errorf("error replaying messags:%v", err)
			}

			if err != nil {
				t.Errorf("error getting channel:%v", err)
			}
			if c != tc.expectedCount {
				t.Errorf("expected count to be [%d] got [%d]", tc.expectedCount, c)
			}

			err = ar.DeleteQueuesAndExchanges(ctx, tc.queueName)
			if err != nil {
				t.Errorf("error deleting queues and exchanges:%v", err)
			}

		})
	}
}
