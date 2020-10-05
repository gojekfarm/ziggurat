package ziggurat

import "context"

type HttpServer interface {
	Start(ctx context.Context, app App)
	Stop() error
}
