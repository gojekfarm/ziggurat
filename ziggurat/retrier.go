package ziggurat

import "context"

type MessageRetrier interface {
	Start(ctx context.Context, applicationContext ApplicationContext) error
	Retry(applicationContext ApplicationContext, payload MessageEvent) error
	Stop() error
	Consume(ctx context.Context, applicationContext ApplicationContext)
	Replay(applicationContext ApplicationContext, topicEntity string, count int) error
}
