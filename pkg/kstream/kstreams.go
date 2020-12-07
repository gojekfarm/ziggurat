package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zlog"
	"strings"
	"sync"
)

type KafkaStreams struct {
	routeConsumerMap map[string][]*kafka.Consumer
}

func NewKafkaStreams() *KafkaStreams {
	return &KafkaStreams{
		routeConsumerMap: map[string][]*kafka.Consumer{},
	}
}

func (k *KafkaStreams) Start(app z.App) (chan struct{}, error) {
	var wg sync.WaitGroup
	config := app.ConfigStore().Config()
	stopChan := make(chan struct{})
	srConfig := config.StreamRouter
	handler := app.Handler()

	for _, route := range app.Routes() {
		streamRouterCfg, ok := srConfig[route]
		if !ok {
			continue
		}
		consumerConfig := NewConsumerConfig(streamRouterCfg.BootstrapServers, streamRouterCfg.GroupID)
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		k.routeConsumerMap[route] = StartConsumers(app, consumerConfig, route, topics, streamRouterCfg.InstanceCount, handler, &wg)
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
