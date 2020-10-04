package ziggurat

import "context"

type MessageRetrier interface {
	Start(ctx context.Context, applicationContext App) error
	Retry(applicationContext App, payload MessageEvent) error
	Stop() error
	Consume(ctx context.Context, applicationContext App)
	Replay(applicationContext App, topicEntity string, count int) error
}
