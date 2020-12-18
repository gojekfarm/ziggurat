package ziggurat

import "context"

type HandlerFunc func(messageEvent MessageEvent, ctx context.Context) ProcessStatus

func (h HandlerFunc) HandleMessage(event MessageEvent, ctx context.Context) ProcessStatus {
	return h(event, ctx)
}
