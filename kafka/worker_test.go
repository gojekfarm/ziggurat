package kafka

import (
	"sync/atomic"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestConsumerWorker_Run(t *testing.T) {
	concurrency := 10
	callCount := int32(0)
	cw := NewWorker(concurrency)
	send, stop := cw.run(func(message *kafka.Message) {
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
