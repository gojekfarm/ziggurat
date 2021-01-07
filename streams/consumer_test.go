package streams

import (
	"bytes"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/rs/zerolog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func TestConsumer_create(t *testing.T) {
	l := logger.NewJSONLogger("disabled")
	cfgMap := NewConsumerConfig("localhost:9092", "bar")
	handler := ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})
	oldStartConsumer := startConsumer
	oldCreateConsumer := createConsumer
	defer func() {
		startConsumer = oldStartConsumer
		createConsumer = oldCreateConsumer
	}()
	startConsumer = func(ctx context.Context, h ziggurat.Handler, l ziggurat.StructuredLogger, consumer *kafka.Consumer, route string, instanceID string, wg *sync.WaitGroup) {

	}
	createConsumer = func(consumerConfig *kafka.ConfigMap, l ziggurat.StructuredLogger, topics []string) *kafka.Consumer {
		return &kafka.Consumer{}
	}

	consumers := StartConsumers(context.Background(), cfgMap, "foo", []string{"bar"}, 2, handler, l, &sync.WaitGroup{})
	if len(consumers) < 2 {
		t.Errorf("could not start consuemrs")
	}
}

func TestConsumer_start(t *testing.T) {
	expectedBytes := []byte("foo")
	l := logger.NewJSONLogger("disabled")
	readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
		t := ""
		return &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &t,
				Partition: 0,
				Offset:    0,
				Metadata:  nil,
				Error:     nil,
			},
			Value:         expectedBytes,
			Key:           expectedBytes,
			Timestamp:     time.Time{},
			TimestampType: 0,
			Opaque:        nil,
			Headers:       nil,
		}, nil
	}
	hf := ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		if bytes.Compare(messageEvent.Value(), expectedBytes) != 0 {
			t.Errorf("expected %s but got %s", expectedBytes, messageEvent.Value())
		}
		return ziggurat.ProcessingSuccess
	})
	c := &kafka.Consumer{}

	storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go func() {
		time.Sleep(1 * time.Second)
		cancelFunc()
	}()
	startConsumer(ctx, hf, l, c, "", "", wg)
	wg.Wait()
}

func TestConsumer_AllBrokersDown(t *testing.T) {
	callCount := 0
	l := logger.NewJSONLogger("disabled")
	readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
		callCount++
		return nil, kafka.NewError(kafka.ErrAllBrokersDown, "", true)
	}

	wg := &sync.WaitGroup{}
	c := &kafka.Consumer{}
	deadlineTime := time.Now().Add(time.Second * 6)
	ctx, cancelFunc := context.WithDeadline(context.Background(), deadlineTime)
	defer cancelFunc()
	h := ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})
	wg.Add(1)
	startConsumer(ctx, h, l, c, "", "", wg)
	wg.Wait()
	if callCount < 1 {
		t.Errorf("expected call count to be atleast 2")
	}
}
