package stream

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/util"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type DefaultRouter struct {
	handlerFunctionMap map[string]z.HandlerFunc
	entityNames        []string
}

func (dr *DefaultRouter) HandleMessage(event basic.MessageEvent, app z.App) z.ProcessStatus {
	return dr.handlerFunctionMap[event.TopicEntity].HandleMessage(event, app)
}

func NewRouter() *DefaultRouter {
	return &DefaultRouter{
		handlerFunctionMap: map[string]z.HandlerFunc{},
		entityNames:        []string{},
	}
}

func (dr *DefaultRouter) HandlerFuncEntityMap() map[string]z.HandlerFunc {
	return dr.handlerFunctionMap
}

func (dr *DefaultRouter) TopicEntities() []string {
	return dr.entityNames
}

func (dr *DefaultRouter) HandlerFunc(topicEntityName string, handlerFn z.HandlerFunc, mw ...z.MiddlewareFunc) {
	dr.entityNames = append(dr.entityNames, topicEntityName)
	if len(mw) > 0 {
		dr.handlerFunctionMap[topicEntityName] = util.PipeHandlers(mw...)(handlerFn)
		return
	}
	dr.handlerFunctionMap[topicEntityName] = handlerFn
}

func (dr *DefaultRouter) Use(middlewareFunc ...z.MiddlewareFunc) {
	for te, handler := range dr.handlerFunctionMap {
		origHandler := handler
		if len(middlewareFunc) > 0 {
			dr.handlerFunctionMap[te] = util.PipeHandlers(middlewareFunc...)(origHandler)
		}
	}
}
