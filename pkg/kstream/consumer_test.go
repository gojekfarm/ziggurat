package kstream

import (
	"bytes"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/handler"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/rs/zerolog"
	"os"
	"sync"
	"testing"
	"time"
)

type consumerTestMockApp struct{}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func (c consumerTestMockApp) Context() context.Context {
	panic("implement me")
}

func (c consumerTestMockApp) Router() z.StreamRouter {
	panic("implement me")
}

func (c consumerTestMockApp) MessageRetry() z.MessageRetry {
	panic("implement me")
}

func (c consumerTestMockApp) Run(router z.StreamRouter, options z.RunOptions) chan struct{} {
	panic("implement me")
}

func (c consumerTestMockApp) Configure(configFunc func(o z.App) z.Options) {
	panic("implement me")
}

func (c consumerTestMockApp) MetricPublisher() z.MetricPublisher {
	panic("implement me")
}

func (c consumerTestMockApp) HTTPServer() z.HttpServer {
	panic("implement me")
}

func (c consumerTestMockApp) Config() *basic.Config {
	panic("implement me")
}

func (c consumerTestMockApp) ConfigReader() z.ConfigReader {
	panic("implement me")
}

func (c consumerTestMockApp) Stop() {
	panic("implement me")
}

func (c consumerTestMockApp) IsRunning() bool {
	panic("implement me")
}

func TestConsumer_create(t *testing.T) {
	ctx := context.Background()
	app := &consumerTestMockApp{}
	cfgMap := kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "myGroup",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "mock",
		"enable.auto.offset.store": false,
	}
	var handler z.HandlerFunc = func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	}
	consumers := StartConsumers(ctx, app, &cfgMap, "test-entity", []string{"plain-text-log"}, 2, handler, &sync.WaitGroup{})
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
	var handlerFunc z.HandlerFunc = func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		if bytes.Compare(messageEvent.MessageValueBytes, expectedBytes) != 0 {
			t.Errorf("expected %s but got %s", expectedBytes, messageEvent.MessageValueBytes)
		}
		return z.ProcessingSuccess
	}
	c := &kafka.Consumer{}
	handler.MessageHandler = func(app z.App, handlerFunc z.HandlerFunc) func(event basic.MessageEvent) {
		return func(event basic.MessageEvent) {
			handlerFunc(event, app)
		}
	}
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
	startConsumer(ctx, app, handlerFunc, c, "", "", wg)
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
	hf := func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	}
	wg.Add(1)
	startConsumer(ctx, app, hf, c, "", "", wg)
	wg.Wait()
	if callCount < 1 {
		t.Errorf("expected call count to be atleast 2")
	}
}
