package ziggurat

import "context"

type MessageRetrier interface {
	Start(config Config, streamRoutes TopicEntityHandlerMap) error
	Retry(config Config, payload MessageEvent) error
	Stop() error
	Consume(ctx context.Context, config Config, streamRoutes TopicEntityHandlerMap)
	Replay(config Config, streamRoutes TopicEntityHandlerMap, topicEntity string, count int)
}
