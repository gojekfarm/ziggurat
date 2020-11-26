package kstream

import (
	"bytes"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/void"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/rs/zerolog"
	"os"
	"sync"
	"testing"
	"time"
)

type consumerTestMockApp struct{}

func (c consumerTestMockApp) Context() context.Context {
	panic("implement me")
}

func (c consumerTestMockApp) Routes() []string {
	panic("implement me")
}

func (c consumerTestMockApp) MessageRetry() z.MessageRetry {
	panic("implement me")
}

func (c consumerTestMockApp) Handler() z.MessageHandler {
	panic("implement me")
}

func (c consumerTestMockApp) MetricPublisher() z.MetricPublisher {
	panic("implement me")
}

func (c consumerTestMockApp) HTTPServer() z.Server {
	panic("implement me")
}

func (c consumerTestMockApp) Config() *basic.Config {
	panic("implement me")
}

func (c consumerTestMockApp) ConfigStore() z.ConfigStore {
	panic("implement me")
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func TestConsumer_create(t *testing.T) {
	ctx := context.Background()
	app := &consumerTestMockApp{}
	cfgMap := NewConsumerConfig("localhost:9092", "bar")
	handler := void.VoidMessageHandler{}
	oldStartConsumer := startConsumer
	oldCreateConsumer := createConsumer
	defer func() {
		startConsumer = oldStartConsumer
		createConsumer = oldCreateConsumer
	}()
	startConsumer = func(ctx context.Context, app z.App, h z.MessageHandler, consumer *kafka.Consumer, topicEntity string, instanceID string, wg *sync.WaitGroup) {
	}
	createConsumer = func(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
		return &kafka.Consumer{}
	}

	consumers := StartConsumers(ctx, app, cfgMap, "foo", []string{"bar"}, 2, handler, &sync.WaitGroup{})
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
	app := &consumerTestMockApp{}
	hf := z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		if bytes.Compare(messageEvent.MessageValueBytes, expectedBytes) != 0 {
			t.Errorf("expected %s but got %s", expectedBytes, messageEvent.MessageValueBytes)
		}
		return z.ProcessingSuccess
	})
	c := &kafka.Consumer{}

	storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancelFunc()
	}()
	startConsumer(ctx, app, hf, c, "", "", wg)
	wg.Wait()
}

func TestConsumer_AllBrokersDown(t *testing.T) {
	callCount := 0
	app := &consumerTestMockApp{}
	readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
		callCount++
		return nil, kafka.NewError(kafka.ErrAllBrokersDown, "", true)
	}

	wg := &sync.WaitGroup{}
	c := &kafka.Consumer{}
	deadlineTime := time.Now().Add(time.Second * 6)
	ctx, cancelFunc := context.WithDeadline(context.Background(), deadlineTime)
	defer cancelFunc()
	h := void.VoidMessageHandler{}
	wg.Add(1)
	startConsumer(ctx, app, h, c, "", "", wg)
	wg.Wait()
	if callCount < 1 {
		t.Errorf("expected call count to be atleast 2")
	}
}
