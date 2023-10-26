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

func newAutoRetry(qn string, count int, consumerCount int, qtype int) *ARetry {
	ar := AutoRetry(
		[]QueueConfig{{
			QueueKey:            qn,
			DelayExpirationInMS: "500",
			RetryCount:          count,
			ConsumerCount:       consumerCount,
			Type:                qtype,
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
			PublishCount:  20,
			RetryCount:    5,
			Name:          "handler is called for PublishCount * RetryCount times",
			QueueName:     "foo",
			ConsumerCount: 1,
		},

		{
			PublishCount:  10,
			RetryCount:    2,
			Name:          "expect handler to be called PublishCount*RetryCount times with multiple consumers",
			QueueName:     "baz",
			ConsumerCount: 25,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx, cfn := context.WithTimeout(context.Background(), 5000*time.Millisecond)
			defer cfn()
			var callCount int32
			expectedCallCount := int32(c.PublishCount * c.RetryCount)
			ar := newAutoRetry(c.QueueName, c.RetryCount, c.ConsumerCount, RetryQueue)
			err := ar.InitPublishers(ctx)
			t.Logf("publishers init successful")
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
			qname:             "foo",
			publishCount:      5,
			viewCount:         5,
			expectedViewCount: 5,
		},
		{
			name:              "read excess number of messages than there are in the queue",
			qname:             "foo",
			publishCount:      5,
			viewCount:         10,
			expectedViewCount: 5,
		},
		{
			name:              "read negative number of messages",
			qname:             "foo",
			publishCount:      5,
			viewCount:         -1,
			expectedViewCount: 0,
		},
		{
			name:              "read zero messages",
			qname:             "foo",
			viewCount:         0,
			publishCount:      5,
			expectedViewCount: 0,
		}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ar := newAutoRetry(c.qname, 5, 1, RetryQueue)
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
			queueName := fmt.Sprintf("%s_%s_%s", c.qname, "dlq", "queue")
			events, err := ar.view(ctx, queueName, c.viewCount, false)
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
			ar := newAutoRetry(tc.queueName, 5, 1, RetryQueue)
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

func Test_MessageLoss(t *testing.T) {
	retryCount := 1000
	consumerCount := 1
	publishCount := 1
	qname := "foo"
	ctx, cfn := context.WithTimeout(context.Background(), time.Second*10)
	defer cfn()
	ar := newAutoRetry(qname, retryCount, consumerCount, RetryQueue)

	done := make(chan struct{})
	go func() {
		err := ar.Stream(ctx, ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
			return ziggurat.Retry
		}, qname))
		if !errors.Is(err, ErrCleanShutdown) {
			t.Errorf("exepcted error to be [%v] got [%v]", ErrCleanShutdown, err)
		}
		done <- struct{}{}
	}()

	err := ar.InitPublishers(ctx)

	if err != nil {
		t.Errorf("publisher init error:%v", err)
	}

	for i := 0; i < publishCount; i++ {
		err := ar.publish(ctx, &ziggurat.Event{
			Value: []byte(fmt.Sprintf("%s_%d", "foo", i)),
		}, qname)
		if err != nil {
			t.Logf("publish error:%v", err)
		}
	}

	<-done
	viewCtx := context.Background()
	time.Sleep(10 * time.Second)

	evts, err := ar.view(viewCtx, fmt.Sprintf("%s_%s_%s", qname, "instant", "queue"), publishCount, false)
	if err != nil {
		t.Errorf("view error:%v", err)
	}

	if len(evts) != publishCount {
		t.Errorf("expected events count to be [%d] but got [%d]", publishCount, len(evts))
	}

	err = ar.DeleteQueuesAndExchanges(viewCtx, qname)
	if err != nil {
		t.Errorf("error deleting queues:%v", err)
	}
}
