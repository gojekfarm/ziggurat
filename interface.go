package ziggurat

import (
	"context"
)

type App interface {
	Context() context.Context
	Routes() Routes
	Handler() MessageHandler
}

type MessageHandler interface {
	HandleMessage(event MessageEvent, app App) ProcessStatus
}

type Streams interface {
	Start(a App) (chan struct{}, error)
	Stop()
}
