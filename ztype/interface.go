package ztype

import (
	"context"
	"github.com/gojekfarm/ziggurat/zbase"
)

type App interface {
	Context() context.Context
	Routes() zbase.Routes
	Handler() MessageHandler
}

type MessageHandler interface {
	HandleMessage(event zbase.MessageEvent, app App) ProcessStatus
}

type Streams interface {
	Start(a App) (chan struct{}, error)
	Stop()
}
