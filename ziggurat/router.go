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

type TopicEntityHandlerMap = map[string]topicEntity

type InstanceCount = map[string]int

type StreamRouter struct {
	handlerFunctionMap map[string]*topicEntity
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9092",
		"group.id":           "myGroup",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}
}

func NewStreamRouter() *StreamRouter {
	return &StreamRouter{
		handlerFunctionMap: make(map[string]*topicEntity),
	}
}

func (sr *StreamRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc) {
	sr.handlerFunctionMap[topicEntityName] = &topicEntity{handlerFunc: handlerFn}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func streamRoutesToMap(streamRoutes []StreamRouterConfig) map[string]StreamRouterConfig {
	streamRouteMap := make(map[string]StreamRouterConfig)
	for _, streamRoute := range streamRoutes {
		streamRouteMap[streamRoute.TopicEntity] = streamRoute
	}
	return streamRouteMap
}

func (sr *StreamRouter) Start(ctx context.Context, config Config) chan int {
	stopNotifierCh := make(chan int)
	var wg sync.WaitGroup
	srConfig := streamRoutesToMap(config.StreamRouters)
	hfMap := sr.handlerFunctionMap
	if len(hfMap) == 0 {
		log.Error().Err(ErrNoHandlersRegistered)
	}

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]
		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := consumerConfig.Set(bootstrapServers); setErr != nil {
			log.Error().Err(setErr)
		}
		if setErr := consumerConfig.Set(groupID); setErr != nil {
			log.Error().Err(setErr)
		}
		go func(te *topicEntity, topicEntityName string, wg *sync.WaitGroup) {
			StartConsumers(ctx, consumerConfig, topicEntityName, strings.Split(streamRouterCfg.OriginTopics, ","), streamRouterCfg.InstanceCount, te.handlerFunc, wg)
			wg.Wait()
			close(stopNotifierCh)
		}(te, topicEntityName, &wg)
	}
	return stopNotifierCh
}
