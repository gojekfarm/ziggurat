package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"
)

type KafkaStreams struct {
	routeConsumerMap map[string][]*kafka.Consumer
	logger           StructuredLogger
	streamRoutes     StreamRoutes
}

func NewKafkaStreams(l StructuredLogger, streamRoutes StreamRoutes) *KafkaStreams {
	return &KafkaStreams{routeConsumerMap: map[string][]*kafka.Consumer{}, logger: l, streamRoutes: streamRoutes}
}

func (k *KafkaStreams) Consume(ctx context.Context, handler Handler) chan error {
	var wg sync.WaitGroup
	stopChan := make(chan error)
	for routeName, stream := range k.streamRoutes {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.ConsumerGroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[routeName] = StartConsumers(ctx, consumerConfig, routeName, topics, stream.ConsumerCount, handler, k.logger, &wg)
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
			k.logger.Error("error stopping consumer %v", consumers[i].Close(), nil)
		}
	}
}
