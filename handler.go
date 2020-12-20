package ziggurat

import "context"

type HandlerFunc func(messageEvent Message, ctx context.Context) ProcessStatus

func (h HandlerFunc) HandleMessage(event Message, ctx context.Context) ProcessStatus {
	return h(event, ctx)
}
