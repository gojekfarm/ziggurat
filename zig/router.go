package zig

import (
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
	entityName       string
}

type Middleware = []MiddlewareFunc

type TopicEntityHandlerMap = map[string]*topicEntity

type DefaultRouter struct {
	handlerFunctionMap TopicEntityHandlerMap
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "myGroup",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "consumer",
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

func (dr *DefaultRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc) {
	dr.handlerFunctionMap[topicEntityName] = &topicEntity{handlerFunc: handlerFn, entityName: topicEntityName}
	if len(mw) > 0 {
		origHandler := dr.handlerFunctionMap[topicEntityName].handlerFunc
		dr.handlerFunctionMap[topicEntityName].handlerFunc = PipeHandlers(mw...)(origHandler)
	}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func (dr *DefaultRouter) stop() {
	for _, te := range dr.GetTopicEntities() {
		log.Info().Str("topic-entity", te.entityName).Msg("stopping consumers")
		for i, _ := range te.consumers {
			if closeErr := te.consumers[i].Close(); closeErr != nil {
				routerLogger.Error().Err(closeErr)
			}
		}
	}
}

func (dr *DefaultRouter) validate(config *Config) {
	srmap := config.StreamRouter
	for entityName, _ := range dr.handlerFunctionMap {
		if _, ok := srmap[entityName]; !ok {
			routerLogger.Warn().Str("registered-route", entityName).Err(ErrInvalidRouteRegistered).Msg("")
		}
	}
}

func (dr *DefaultRouter) Start(app *App) (chan int, error) {
	ctx := app.Context()
	config := app.config
	stopChan := make(chan int)
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := dr.handlerFunctionMap
	if len(hfMap) == 0 {
		routerLogger.Fatal().Err(ErrNoHandlersRegistered).Msg("")
	}

	dr.validate(app.config)
	// halts app if validation fails

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]

		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := consumerConfig.Set(bootstrapServers); setErr != nil {
			routerLogger.Error().Err(setErr).Msg("")
			return nil, setErr
		}
		if setErr := consumerConfig.Set(groupID); setErr != nil {
			routerLogger.Error().Err(setErr).Msg("")
			return nil, setErr
		}
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		consumers := startConsumers(ctx, app, consumerConfig, topicEntityName, topics, streamRouterCfg.InstanceCount, te.handlerFunc, &wg)
		te.consumers = consumers
	}

	go func() {
		wg.Wait()
		dr.stop()
		close(stopChan)
	}()

	return stopChan, nil
}
