package ziggurat

import (
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

func (k *KafkaStreams) Start(z *Ziggurat) (chan struct{}, error) {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	handler := z.Handler()

	for routeName, stream := range z.Routes() {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.GroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[routeName] = StartConsumers(z, consumerConfig, routeName, topics, stream.InstanceCount, handler, &wg)
	}

	go func() {
		wg.Wait()
		k.Stop()
		close(stopChan)
	}()

	return stopChan, nil
}

func (k *KafkaStreams) Stop() {
	for _, consumers := range k.routeConsumerMap {
		for i, _ := range consumers {
			LogError(consumers[i].Close(), "consumer close error", nil)
		}
	}
}
