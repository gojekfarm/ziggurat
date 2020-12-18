package ziggurat

import (
	"bytes"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
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
	app := NewApp()
	cfgMap := NewConsumerConfig("localhost:9092", "bar")
	handler := HandlerFunc(func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus {
		return ProcessingSuccess
	})
	oldStartConsumer := startConsumer
	oldCreateConsumer := createConsumer
	defer func() {
		startConsumer = oldStartConsumer
		createConsumer = oldCreateConsumer
	}()
	startConsumer = func(z *Ziggurat, h MessageHandler, consumer *kafka.Consumer, topicEntity string, instanceID string, wg *sync.WaitGroup) {
	}
	createConsumer = func(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
		return &kafka.Consumer{}
	}

	consumers := StartConsumers(app, cfgMap, "foo", []string{"bar"}, 2, handler, &sync.WaitGroup{})
	if len(consumers) < 2 {
		t.Errorf("could not start consuemrs")
	}
}

func TestConsumer_start(t *testing.T) {
	expectedBytes := []byte("foo")
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
	app := NewApp()
	hf := HandlerFunc(func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus {
		if bytes.Compare(messageEvent.MessageValueBytes, expectedBytes) != 0 {
			t.Errorf("expected %s but got %s", expectedBytes, messageEvent.MessageValueBytes)
		}
		return ProcessingSuccess
	})
	c := &kafka.Consumer{}

	storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	app.ctx = ctx
	defer cancelFunc()
	go func() {
		time.Sleep(1 * time.Second)
		cancelFunc()
	}()
	startConsumer(app, hf, c, "", "", wg)
	wg.Wait()
}

func TestConsumer_AllBrokersDown(t *testing.T) {
	callCount := 0
	app := NewApp()
	readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
		callCount++
		return nil, kafka.NewError(kafka.ErrAllBrokersDown, "", true)
	}

	wg := &sync.WaitGroup{}
	c := &kafka.Consumer{}
	deadlineTime := time.Now().Add(time.Second * 6)
	ctx, cancelFunc := context.WithDeadline(context.Background(), deadlineTime)
	defer cancelFunc()
	app.ctx = ctx
	h := HandlerFunc(func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus {
		return ProcessingSuccess
	})
	wg.Add(1)
	startConsumer(app, h, c, "", "", wg)
	wg.Wait()
	if callCount < 1 {
		t.Errorf("expected call count to be atleast 2")
	}
}
