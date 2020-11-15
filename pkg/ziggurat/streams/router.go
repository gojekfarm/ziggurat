package streams

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/cons"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/util"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/z"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/zerror"
	"strings"
	"sync"
)

type DefaultRouter struct {
	handlerFunctionMap z.TopicEntityHandlerMap
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
		handlerFunctionMap: make(map[string]*z.TopicEntity),
	}
}

func (dr *DefaultRouter) GetHandlerFunctionMap() map[string]*z.TopicEntity {
	return dr.handlerFunctionMap
}

func (dr *DefaultRouter) GetTopicEntities() []*z.TopicEntity {
	var topicEntities []*z.TopicEntity
	for _, te := range dr.handlerFunctionMap {
		topicEntities = append(topicEntities, te)
	}
	return topicEntities
}

func (dr *DefaultRouter) GetTopicEntityNames() []string {
	tes := dr.GetTopicEntities()
	var names []string
	for _, te := range tes {
		names = append(names, te.EntityName)
	}
	return names
}

func (dr *DefaultRouter) HandlerFunc(topicEntityName string, handlerFn z.HandlerFunc, mw ...z.MiddlewareFunc) {
	dr.handlerFunctionMap[topicEntityName] = &z.TopicEntity{HandlerFunc: handlerFn, EntityName: topicEntityName}
	if len(mw) > 0 {
		origHandler := dr.handlerFunctionMap[topicEntityName].HandlerFunc
		dr.handlerFunctionMap[topicEntityName].HandlerFunc = util.PipeHandlers(mw...)(origHandler)
	}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func (dr *DefaultRouter) stop() {
	for _, te := range dr.GetTopicEntities() {
		logger.LogInfo("stopping consumers", map[string]interface{}{"topic-entity": te.EntityName})
		for i, _ := range te.Consumers {
			logger.LogError(te.Consumers[i].Close(), "consumer close error", nil)
		}
	}
}

func (dr *DefaultRouter) validate(config *basic.Config) {
	srmap := config.StreamRouter
	for entityName, _ := range dr.handlerFunctionMap {
		if _, ok := srmap[entityName]; !ok {
			args := map[string]interface{}{"invalid-entity-name": entityName}
			logger.LogWarn("router: registered route not found in config", args)
		}
	}
}

func (dr *DefaultRouter) Start(app z.App) (chan int, error) {
	ctx := app.Context()
	config := app.Config()
	stopChan := make(chan int)
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := dr.handlerFunctionMap
	if len(hfMap) == 0 {
		logger.LogFatal(zerror.ErrNoHandlersRegistered, "ziggurat router", nil)
	}

	dr.validate(config)

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]

		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := setConsumerConfig(consumerConfig, bootstrapServers); setErr != nil {
			logger.LogError(setErr, "ziggurat router", nil)
			return nil, setErr
		}
		if setErr := setConsumerConfig(consumerConfig, groupID); setErr != nil {
			logger.LogError(setErr, "ziggurat router", nil)
			return nil, setErr
		}
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		consumers := cons.StartConsumers(ctx, app, consumerConfig, topicEntityName, topics, streamRouterCfg.InstanceCount, te.HandlerFunc, &wg)
		te.Consumers = consumers
	}

	go func() {
		wg.Wait()
		dr.stop()
		close(stopChan)
	}()

	return stopChan, nil

}
