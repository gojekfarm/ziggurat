package kafka

import (
	"context"
	"sync"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

func TestKafkaStreams_Consume(t *testing.T) {
	routes := StreamConfig{{RouteGroup: "foo"}}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(ctx context.Context, consumerConfig *kafka.ConfigMap, route string, topics []string, instances int, h ziggurat.Handler, l ziggurat.StructuredLogger, wg *sync.WaitGroup) []*kafka.Consumer {
		return []*kafka.Consumer{}
	}
	ks := Streams{
		routeConsumerMap: nil,
		Logger:           logger.NOOP,
		StreamConfig:     StreamConfig{{RouteGroup: "foo"}},
	}
	f := func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	}
	ks.Stream(context.Background(), ziggurat.HandlerFunc(f))
	if len(ks.routeConsumerMap) < len(routes) {
		t.Errorf("expected count %d but got %d", len(ks.routeConsumerMap), len(routes))
	}
}
