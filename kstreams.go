package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"
)

type RouteGroup struct {
	BootstrapServers string
	OriginTopics     string
	ConsumerGroupID  string
	ConsumerCount    int
}

type KafkaRouteGroup map[string]RouteGroup

type KafkaStreams struct {
	routeConsumerMap map[string][]*kafka.Consumer
	Logger           StructuredLogger
	KafkaRouteGroup
}

func (k *KafkaStreams) Consume(ctx context.Context, handler Handler) chan error {
	if k.Logger == nil {
		k.Logger = NewLogger("info")
	}
	var wg sync.WaitGroup
	k.routeConsumerMap = map[string][]*kafka.Consumer{}
	stopChan := make(chan error)
	for routeName, stream := range k.KafkaRouteGroup {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.ConsumerGroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[routeName] = StartConsumers(ctx, consumerConfig, routeName, topics, stream.ConsumerCount, handler, k.Logger, &wg)
	}

	go func() {
		wg.Wait()
		k.stop()
		stopChan <- nil
	}()

	return stopChan
}

func (k *KafkaStreams) stop() {
	for _, consumers := range k.routeConsumerMap {
		for i, _ := range consumers {
			k.Logger.Error("error stopping consumer %v", consumers[i].Close(), nil)
		}
	}
}
