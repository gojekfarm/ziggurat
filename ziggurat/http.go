package ziggurat

import "context"

type HttpServer interface {
	Start(ctx context.Context, applicationContext App)
	Stop() error
}
