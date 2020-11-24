package stream

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/util"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type defaultRouter struct {
	handlerFunctionMap map[string]z.HandlerFunc
}

func (dr *defaultRouter) HandleMessage(event basic.MessageEvent, app z.App) z.ProcessStatus {
	route := event.StreamRoute
	return dr.handlerFunctionMap[route](event, app)
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]z.HandlerFunc{},
	}
}

func (dr *defaultRouter) HandlerFunc(route string, handlerFunc func(event basic.MessageEvent, app z.App) z.ProcessStatus) {
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...z.MiddlewareFunc) z.MessageHandler {
	return util.PipeHandlers(mw...)(dr)
}
