package kstream

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat/mock"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestConsumerWorker_Run(t *testing.T) {
	concurrency := 10
	callCount := int32(0)
	cw := NewWorker(concurrency)
	send, stop := cw.run(mock.NewZig(), func(message *kafka.Message) {
		atomic.AddInt32(&callCount, 1)
	})
	for i := 0; i < concurrency; i++ {
		send <- &kafka.Message{}
	}
	close(send)
	<-stop
	if callCount != int32(concurrency) {
		t.Errorf("expectec call count %d but got %d", concurrency, callCount)
	}
}

func TestConsumerWorker_ContextDone(t *testing.T) {
	concurrency := 10
	jobs := concurrency + 20
	callCount := int32(0)
	cw := NewWorker(concurrency)
	a := mock.NewZig()
	c, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancelFunc()
	a.ContextFunc = func() context.Context {
		return c
	}

	send, stop := cw.run(a, func(message *kafka.Message) {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&callCount, 1)
	})

	go func() {
		for i := 0; i < jobs; i++ {
			send <- &kafka.Message{}
		}
	}()
	<-stop
	if !strings.Contains(c.Err().Error(), "deadline exceeded") {
		t.Errorf("exepcted context deadline to be cancelled")
	}
}
