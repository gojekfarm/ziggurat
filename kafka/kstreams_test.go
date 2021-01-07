package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"sync"
	"testing"
)

func TestKafkaStreams_Consume(t *testing.T) {
	routes := RouteGroup{"foo": {}}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(ctx context.Context, consumerConfig *kafka.ConfigMap, route string, topics []string, instances int, h ziggurat.Handler, l ziggurat.StructuredLogger, wg *sync.WaitGroup) []*kafka.Consumer {
		return []*kafka.Consumer{}
	}
	kstreams := Streams{
		routeConsumerMap: nil,
		Logger:           logger.NewJSONLogger("disabled"),
		RouteGroup:       RouteGroup{"foo": {}},
	}
	kstreams.Stream(context.Background(), ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	}))
	if len(kstreams.routeConsumerMap) < len(routes) {
		t.Errorf("expected count %d but got %d", len(kstreams.routeConsumerMap), len(routes))
	}
}
