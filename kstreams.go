package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mitchellh/mapstructure"
	"strings"
	"sync"
)

const (
	KafkaKeyBootstrapServers = "bootstrap-servers"
	KafkaKeyConsumerGroup    = "group-id"
	KafkaKeyOriginTopics     = "origin-topics"
	KafkaKeyConsumerCount    = "consumer-count"
)

type streamConfig struct {
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	GroupID          string `mapstructure:"group-id"`
	OriginTopics     string `mapstructure:"origin-topics"`
	ConsumerCount    int    `mapstructure:"consumer-count"`
}

type KafkaStreams struct {
	routeConsumerMap map[string][]*kafka.Consumer
	l                StructuredLogger
	consumerConf     map[string]streamConfig
}

func NewKafkaStreams(l StructuredLogger, consumerConf Routes) *KafkaStreams {
	streamConfig := &map[string]streamConfig{}
	if err := mapstructure.Decode(consumerConf, streamConfig); err != nil {
		panic(fmt.Errorf("error creating kafka streams: %s", err.Error()))
	}
	return &KafkaStreams{routeConsumerMap: map[string][]*kafka.Consumer{}, l: l, consumerConf: *streamConfig}
}

func (k *KafkaStreams) Consume(ctx context.Context, handler Handler) chan error {
	var wg sync.WaitGroup
	stopChan := make(chan error)
	for routeName, stream := range k.consumerConf {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.GroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		k.routeConsumerMap[routeName] = StartConsumers(ctx, consumerConfig, routeName, topics, stream.ConsumerCount, handler, k.l, &wg)
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
			k.l.Error("error stopping consumer %v", consumers[i].Close(), nil)
		}
	}
}
