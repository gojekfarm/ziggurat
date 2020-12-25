package ziggurat

import "fmt"

type defaultRouter struct {
	handlerFunctionMap map[string]HandlerFunc
	l                  StructuredLogger
}

func (dr *defaultRouter) HandleEvent(event Event, ) ProcessStatus {
	route := event.Header(HeaderMessageRoute)
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		dr.l.Warn("handler not found", map[string]interface{}{"routing-key": route})
		return SkipMessage
	} else {
		return handler.HandleEvent(event)
	}
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]HandlerFunc{},
		l:                  NewLogger("info"),
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event Event) ProcessStatus) {
	if route == "" {
		panic(`route cannot be ""`)
	}
	if handlerFunc == nil {
		panic("handler cannot be nil")
	}

	if _, ok := dr.handlerFunctionMap[route]; ok {
		panic(fmt.Sprintf("route %s has already been registered", route))
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...func(h Handler) Handler) Handler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Lookup(route string) (Handler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
