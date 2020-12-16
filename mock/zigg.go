package mock

import (
	"context"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

type Zig struct {
	ContextFunc       func() context.Context
	RoutesFunc        func() zbase.Routes
	HandlerFuncGetter func() ztype.MessageHandler
}

func NewZig() *Zig {
	return &Zig{
		ContextFunc: func() context.Context {
			return context.Background()
		},
		RoutesFunc: func() zbase.Routes {
			return zbase.Routes{}
		},
		HandlerFuncGetter: func() ztype.MessageHandler {
			return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
				return ztype.ProcessingSuccess
			})
		},
	}
}

func (z Zig) Context() context.Context {
	return z.ContextFunc()
}

func (z Zig) Routes() zbase.Routes {
	return z.RoutesFunc()
}

func (z Zig) Handler() ztype.MessageHandler {
	return z.HandlerFuncGetter()
}
