package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
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

func (k *KafkaStreams) Start(app ztype.App) (chan struct{}, error) {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	handler := app.Handler()

	for _, stream := range app.Routes() {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.GroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[stream.RouteName] = StartConsumers(app, consumerConfig, stream.RouteName, topics, stream.InstanceCount, handler, &wg)
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
			zlog.LogError(consumers[i].Close(), "consumer close error", nil)
		}
	}
}
