package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"strings"
	"sync"
)

type KafkaStreams struct {
	entityConsumerMap map[string][]*kafka.Consumer
}

func NewKafkaStreams() *KafkaStreams {
	return &KafkaStreams{
		entityConsumerMap: map[string][]*kafka.Consumer{},
	}
}

func (k *KafkaStreams) Start(app z.App) (chan struct{}, error) {
	var wg sync.WaitGroup
	ctx := app.Context()
	config := app.Config()
	stopChan := make(chan struct{})
	srConfig := config.StreamRouter
	hfMap := app.Router().RouteHandlerMap()

	if len(hfMap) == 0 {
		logger.LogFatal(zerror.ErrNoHandlersRegistered, "kafka streams", nil)
	}

	for te, h := range hfMap {
		streamRouterCfg := srConfig[te]
		consumerConfig := NewConsumerConfig(streamRouterCfg.BootstrapServers, streamRouterCfg.GroupID)
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		k.entityConsumerMap[te] = StartConsumers(ctx, app, consumerConfig, te, topics, streamRouterCfg.InstanceCount, h, &wg)
	}

	go func() {
		wg.Wait()
		k.Stop()
		close(stopChan)
	}()

	return stopChan, nil
}

func (k *KafkaStreams) Stop() {
	for _, consumers := range k.entityConsumerMap {
		for i, _ := range consumers {
			logger.LogError(consumers[i].Close(), "consumer close error", nil)
		}
	}
}
