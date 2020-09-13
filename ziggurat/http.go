package ziggurat

import "context"

type HttpServer interface {
	Start(ctx context.Context, config Config, retrier MessageRetrier, handlerFuncMap TopicEntityHandlerMap)
	Stop() error
}
