package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
	"testing"
)

func TestKafkaStreams_Start(t *testing.T) {
	l := NewLogger("disabled")
	routes := StreamRoutes{"foo": {}}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(ctx context.Context, consumerConfig *kafka.ConfigMap, route string, topics []string, instances int, h Handler, l StructuredLogger, wg *sync.WaitGroup) []*kafka.Consumer {
		return []*kafka.Consumer{}
	}
	kstreams := NewKafkaStreams(l, StreamRoutes{"foo": {}})
	kstreams.Consume(context.Background(), HandlerFunc(func(messageEvent Message, ctx context.Context) ProcessStatus {
		return ProcessingSuccess
	}))
	if len(kstreams.routeConsumerMap) < len(routes) {
		t.Errorf("expected count %d but got %d", len(kstreams.routeConsumerMap), len(routes))
	}

}
