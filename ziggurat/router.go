package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
)

type topicEntity struct {
	handlerFunc      HandlerFunc
	consumers        []*kafka.Consumer
	bootstrapServers string
	originTopics     []string
}

type TopicEntityHandlerMap = map[string]*topicEntity

type InstanceCount = map[string]int

type StreamRouter struct {
	handlerFunctionMap TopicEntityHandlerMap
	messageRetrier     MessageRetrier
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":       "localhost:9092",
		"group.id":                "myGroup",
		"auto.offset.reset":       "earliest",
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 2000,
		"debug":                   "consumer,broker",
	}
}

func NewStreamRouter() *StreamRouter {
	return &StreamRouter{
		handlerFunctionMap: make(map[string]*topicEntity),
	}
}

func (sr *StreamRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return sr.handlerFunctionMap
}

func (sr *StreamRouter) GetTopicEntities() []*topicEntity {
	var topicEntities []*topicEntity
	for _, te := range sr.handlerFunctionMap {
		topicEntities = append(topicEntities, te)
	}
	return topicEntities
}

func (sr *StreamRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc) {
	sr.handlerFunctionMap[topicEntityName] = &topicEntity{handlerFunc: handlerFn}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func notifyRouterStop(stopChannel chan<- int, wg *sync.WaitGroup) {
	wg.Wait()
	close(stopChannel)
}

func (sr *StreamRouter) Start(ctx context.Context, config Config, retrier MessageRetrier) chan int {
	stopNotifierCh := make(chan int)
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := sr.handlerFunctionMap
	if len(hfMap) == 0 {
		log.Fatal().Err(ErrNoHandlersRegistered).Msg("router error")
	}

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]
		if topicEntityName != streamRouterCfg.TopicEntity {
			log.Fatal().Err(ErrTopicEntityMismatch).Msg("router error")
		}
		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := consumerConfig.Set(bootstrapServers); setErr != nil {
			log.Error().Err(setErr)
		}
		if setErr := consumerConfig.Set(groupID); setErr != nil {
			log.Error().Err(setErr)
		}
		consumers := StartConsumers(ctx, config, consumerConfig, topicEntityName, strings.Split(streamRouterCfg.OriginTopics, ","), streamRouterCfg.InstanceCount, te.handlerFunc, retrier, &wg)
		te.consumers = consumers
	}

	if config.Retry.Enabled {
		log.Info().Msg("starting retrier...")
		if retrierStartErr := retrier.Start(config, hfMap); retrierStartErr != nil {
			log.Fatal().Err(retrierStartErr).Msg("unable to start retrier")
		}

		retrier.Consume(ctx, config, hfMap)
		log.Info().Msg("starting retrier consumer")
	}

	go notifyRouterStop(stopNotifierCh, &wg)

	return stopNotifierCh
}
