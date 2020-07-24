package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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

func (sr *StreamRouter) Start() {
	config, _ := ParseConfig()
	srConfig := streamRoutesToMap(config.StreamRouters)
	hfMap := sr.handlerFunctionMap
	if len(hfMap) == 0 {
		log.Fatal("error: unable to start stream-router, no handler functions registered")
	}

	routerCtx, cancelFn := context.WithCancel(context.Background())
	intrCh := make(chan os.Signal)
	signal.Notify(intrCh, os.Interrupt, syscall.SIGTERM)

	go func(intrCh chan os.Signal) {
		<-intrCh
		log.Printf("Terminating app, CTRL+C Interrupt recevied")
		cancelFn()
	}(intrCh)

	var wg sync.WaitGroup
	var consumers []*kafka.Consumer
	for topicEntityName, topicEntity := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]
		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		consumerConfig.Set(bootstrapServers)
		consumerConfig.Set(groupID)
		fmt.Println(consumerConfig)
		consumers = StartConsumers(routerCtx, consumerConfig, topicEntityName, strings.Split(streamRouterCfg.OriginTopics, ","), streamRouterCfg.InstanceCount, topicEntity.handlerFunc, &wg)
	}
	wg.Wait()
	for i, _ := range consumers {
		closeErr := consumers[i].Close()
		if closeErr != nil {
			fmt.Printf("error closing consumer: %v\n", closeErr)
		}
	}
}
