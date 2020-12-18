package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"
)

type KafkaStreams struct {
	routeConsumerMap map[string][]*kafka.Consumer
}

func New() *KafkaStreams {
	return &KafkaStreams{
		routeConsumerMap: map[string][]*kafka.Consumer{},
	}
}

func (k *KafkaStreams) Consume(ctx context.Context, routes Routes, handler MessageHandler) chan error {
	var wg sync.WaitGroup
	stopChan := make(chan error)

	for routeName, stream := range routes {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.GroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[routeName] = StartConsumers(ctx, consumerConfig, routeName, topics, stream.InstanceCount, handler, &wg)
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
			LogError(consumers[i].Close(), "consumer close error", nil)
		}
	}
}
