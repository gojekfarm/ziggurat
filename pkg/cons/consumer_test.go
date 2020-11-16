package cons

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"sync"
	"testing"
)

type consumerTestMockApp struct{}

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
