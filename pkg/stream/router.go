package stream

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/cons"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/util"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"strings"
	"sync"
)

type DefaultRouter struct {
	handlerFunctionMap z.TopicEntityHandlerMap
	routerMiddleware   []z.MiddlewareFunc
}

var setConsumerConfig = func(consumerConfigMap *kafka.ConfigMap, kv string) error {
	return consumerConfigMap.Set(kv)
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
	dr.handlerFunctionMap[topicEntityName] = &z.TopicEntity{HandlerFunc: handlerFn, EntityName: topicEntityName, Middleware: mw}

}

func (dr *DefaultRouter) Use(middlewareFunc z.MiddlewareFunc) {
	dr.routerMiddleware = append(dr.routerMiddleware, middlewareFunc)
}

func (dr *DefaultRouter) attachMiddleware() {
	for _, topicEntity := range dr.handlerFunctionMap {
		allMiddleware := []z.MiddlewareFunc{}
		allMiddleware = append(allMiddleware, dr.routerMiddleware...)
		allMiddleware = append(allMiddleware, topicEntity.Middleware...)
		origHandler := topicEntity.HandlerFunc
		if len(allMiddleware) > 0 {
			topicEntity.HandlerFunc = util.PipeHandlers(allMiddleware...)(origHandler)
		}
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

func (dr *DefaultRouter) Start(app z.App) (chan struct{}, error) {
	ctx := app.Context()
	config := app.Config()
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := dr.handlerFunctionMap
	if len(hfMap) == 0 {
		logger.LogFatal(zerror.ErrNoHandlersRegistered, "ziggurat router", nil)
	}

	dr.validate(config)
	dr.attachMiddleware()

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]

		consumerConfig := cons.NewConsumerConfig()
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
