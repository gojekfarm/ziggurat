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

func (k *KafkaStreams) Start(app AppContext) (chan struct{}, error) {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	handler := app.Handler()

	for routeName, stream := range app.Routes() {
		consumerConfig := NewConsumerConfig(stream.Servers(), stream.ConsumerGroupID())
		topics := strings.Split(stream.Topics(), ",")
		k.routeConsumerMap[routeName] = StartConsumers(app, consumerConfig, routeName, topics, stream.ThreadCount(), handler, &wg)
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
