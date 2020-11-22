package stream

import (
	"fmt"
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
	entityNames        []string
}

func NewRouter() *DefaultRouter {
	return &DefaultRouter{
		handlerFunctionMap: map[string]*z.TopicEntity{},
		routerMiddleware:   []z.MiddlewareFunc{},
		entityNames:        []string{},
	}
}

func (dr *DefaultRouter) HandlerFuncEntityMap() map[string]z.HandlerFunc {
	result := map[string]z.HandlerFunc{}
	for entity, handler := range dr.handlerFunctionMap {
		result[entity] = handler.HandlerFunc
	}
	return result
}

func (dr *DefaultRouter) TopicEntities() []string {
	return dr.entityNames
}

func (dr *DefaultRouter) HandlerFunc(topicEntityName string, handlerFn z.HandlerFunc, mw ...z.MiddlewareFunc) {
	dr.entityNames = append(dr.entityNames, topicEntityName)
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
	for _, te := range dr.handlerFunctionMap {
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
		consumerConfig := cons.NewConsumerConfig(streamRouterCfg.BootstrapServers, streamRouterCfg.GroupID)
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
