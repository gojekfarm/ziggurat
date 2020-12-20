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
}

func NewKafkaStreams(l StructuredLogger) *KafkaStreams {
	return &KafkaStreams{routeConsumerMap: map[string][]*kafka.Consumer{}, l: l}
}

func (k *KafkaStreams) Consume(ctx context.Context, handler Handler, routes Routes) chan error {
	var wg sync.WaitGroup
	stopChan := make(chan error)

	for routeName, rawConfig := range routes {
		stream := &streamConfig{}
		if err := mapstructure.Decode(rawConfig, stream); err != nil {
			go func() {
				stopChan <- fmt.Errorf("consumer config parse error: %v", err.Error())
			}()
			return stopChan
		}
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
