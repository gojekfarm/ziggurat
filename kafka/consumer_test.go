package kafka

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/rs/zerolog"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func TestConsumer_create(t *testing.T) {

	l := logger.NOOP
	cfgMap := NewConsumerConfig("localhost:9092", "bar")
	handler := ziggurat.HandlerFunc(func(messagectx context.Context, event *ziggurat.Event) error {
		return nil
	})
	oldStartConsumer := startConsumer
	oldCreateConsumer := createConsumer
	defer func() {
		startConsumer = oldStartConsumer
		createConsumer = oldCreateConsumer
	}()

	type test struct {
		Name          string
		ExpectedCount int
	}

	tests := []test{{
		Name:          "consumer creation",
		ExpectedCount: 5,
	}}

	startConsumer = func(_ context.Context, _ ziggurat.Handler, _ ziggurat.StructuredLogger, _ *kafka.Consumer, _ string, _ kafka.ConfigMap, _ *sync.WaitGroup) {
	}
	createConsumer = func(_ *kafka.ConfigMap, _ ziggurat.StructuredLogger, _ []string) *kafka.Consumer {
		return &kafka.Consumer{}
	}

	for _, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			consumers := StartConsumers(context.Background(), cfgMap, "foo", []string{"bar"}, c.ExpectedCount, handler, l, &sync.WaitGroup{})
			if len(consumers) != c.ExpectedCount {
				t.Errorf("expected count [%d] got [%d]", c.ExpectedCount, len(consumers))
			}
		})
	}
}

func TestConsumer_start(t *testing.T) {
	expectedBytes := []byte("foo")
	l := logger.NOOP
	var oldPoll = pollEvent
	defer func() {
		pollEvent = oldPoll
	}()
	pollEvent = func(c *kafka.Consumer, pollTimeout int) kafka.Event {
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
		}
	}
	hf := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
		if bytes.Compare(event.Value, expectedBytes) != 0 {
			t.Errorf("expected %s but got %s", expectedBytes, event.Value)
		}
		return nil
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
	startConsumer(ctx, hf, l, c, "", kafka.ConfigMap{}, wg)
	wg.Wait()
}
