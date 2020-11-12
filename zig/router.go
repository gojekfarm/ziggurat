package zig

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"
)

type topicEntity struct {
	HandlerFunc      HandlerFunc
	consumers        []*kafka.Consumer
	bootstrapServers string
	originTopics     []string
	entityName       string
}

type Middleware = []MiddlewareFunc

type TopicEntityHandlerMap = map[string]*topicEntity

type DefaultRouter struct {
	handlerFunctionMap TopicEntityHandlerMap
}

var setConsumerConfig = func(consumerConfigMap *kafka.ConfigMap, kv string) error {
	return consumerConfigMap.Set(kv)
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "myGroup",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "consumer,broker",
		"enable.auto.offset.store": false,
	}
}

func NewRouter() *DefaultRouter {
	return &DefaultRouter{
		handlerFunctionMap: make(map[string]*topicEntity),
	}
}

func (dr *DefaultRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return dr.handlerFunctionMap
}

func (dr *DefaultRouter) GetTopicEntities() []*topicEntity {
	var topicEntities []*topicEntity
	for _, te := range dr.handlerFunctionMap {
		topicEntities = append(topicEntities, te)
	}
	return topicEntities
}

func (dr *DefaultRouter) GetTopicEntityNames() []string {
	tes := dr.GetTopicEntities()
	var names []string
	for _, te := range tes {
		names = append(names, te.entityName)
	}
	return names
}

func (dr *DefaultRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc) {
	dr.handlerFunctionMap[topicEntityName] = &topicEntity{HandlerFunc: handlerFn, entityName: topicEntityName}
	if len(mw) > 0 {
		origHandler := dr.handlerFunctionMap[topicEntityName].HandlerFunc
		dr.handlerFunctionMap[topicEntityName].HandlerFunc = pipeHandlers(mw...)(origHandler)
	}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func (dr *DefaultRouter) stop() {
	for _, te := range dr.GetTopicEntities() {
		logInfo("stopping consumers", map[string]interface{}{"topic-entity": te.entityName})
		for i, _ := range te.consumers {
			logError(te.consumers[i].Close(), "consumer close error", nil)
		}
	}
}

func (dr *DefaultRouter) validate(config *Config) {
	srmap := config.StreamRouter
	for entityName, _ := range dr.handlerFunctionMap {
		if _, ok := srmap[entityName]; !ok {
			args := map[string]interface{}{"invalid-entity-name": entityName}
			logWarn("router: registered route not found in config", args)
		}
	}
}

func (dr *DefaultRouter) Start(app App) (chan int, error) {
	ctx := app.Context()
	config := app.Config()
	stopChan := make(chan int)
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := dr.handlerFunctionMap
	if len(hfMap) == 0 {
		logFatal(ErrNoHandlersRegistered, "ziggurat router", nil)
	}

	dr.validate(config)

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]

		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := setConsumerConfig(consumerConfig, bootstrapServers); setErr != nil {
			logError(setErr, "ziggurat router", nil)
			return nil, setErr
		}
		if setErr := setConsumerConfig(consumerConfig, groupID); setErr != nil {
			logError(setErr, "ziggurat router", nil)
			return nil, setErr
		}
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		consumers := startConsumers(ctx, app, consumerConfig, topicEntityName, topics, streamRouterCfg.InstanceCount, te.HandlerFunc, &wg)
		te.consumers = consumers
	}

	go func() {
		wg.Wait()
		dr.stop()
		close(stopChan)
	}()

	return stopChan, nil

}
