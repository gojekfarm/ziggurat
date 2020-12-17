package ziggurat

import (
	"context"
)

type Zig struct {
	ContextFunc       func() context.Context
	RoutesFunc        func() Routes
	HandlerFuncGetter func() MessageHandler
}

func NewZig() *Zig {
	return &Zig{
		ContextFunc: func() context.Context {
			return context.Background()
		},
		RoutesFunc: func() Routes {
			return Routes{}
		},
		HandlerFuncGetter: func() MessageHandler {
			return HandlerFunc(func(messageEvent MessageEvent, app AppContext) ProcessStatus {
				return ProcessingSuccess
			})
		},
	}
}

func (z Zig) Context() context.Context {
	return z.ContextFunc()
}

func (z Zig) Routes() Routes {
	return z.RoutesFunc()
}

func (z Zig) Handler() MessageHandler {
	return z.HandlerFuncGetter()
}
