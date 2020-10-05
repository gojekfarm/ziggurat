package ziggurat

import "context"

type MessageRetrier interface {
	Start(ctx context.Context, app App) error
	Retry(app App, payload MessageEvent) error
	Stop() error
	Consume(ctx context.Context, app App)
	Replay(app App, topicEntity string, count int) error
}
