package kstream

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/void"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"sync"
	"testing"
)

type kstreamMockApp struct{}

func (k kstreamMockApp) Context() context.Context {
	return context.Background()
}

func (k kstreamMockApp) Routes() []string {
	return []string{"foo", "bar"}
}

func (k kstreamMockApp) MessageRetry() z.MessageRetry {
	panic("implement me")
}

func (k kstreamMockApp) Handler() z.MessageHandler {
	return void.VoidMessageHandler{}
}

func (k kstreamMockApp) MetricPublisher() z.MetricPublisher {
	panic("implement me")
}

func (k kstreamMockApp) HTTPServer() z.HttpServer {
	panic("implement me")
}

func (k kstreamMockApp) Config() *basic.Config {
	return &basic.Config{
		StreamRouter: map[string]basic.StreamRouterConfig{
			"foo": {
				InstanceCount:    0,
				BootstrapServers: "",
				OriginTopics:     "",
				GroupID:          "",
			},
			"bar": {
				InstanceCount:    0,
				BootstrapServers: "",
				OriginTopics:     "",
				GroupID:          "",
			},
		},
	}
}

func (k kstreamMockApp) ConfigStore() z.ConfigStore {
	panic("implement me")
}

func TestKafkaStreams_Start(t *testing.T) {
	a := &kstreamMockApp{}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(routerCtx context.Context, app z.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, h z.MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
		wg.Add(1)
		wg.Done()
		return []*kafka.Consumer{}
	}
	kstreams := NewKafkaStreams()
	done, _ := kstreams.Start(a)
	if len(kstreams.routeConsumerMap) < len(a.Routes()) {
		t.Errorf("expected count %d but got %d", len(kstreams.routeConsumerMap), len(a.Routes()))
	}
	<-done

}
