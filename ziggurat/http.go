package ziggurat

import "context"

type Http interface {
	Start(ctx context.Context, config Config, retrier MessageRetrier, handlerFuncMap TopicEntityHandlerMap)
	Stop() error
}
