package ziggurat

import "context"

type HttpServer interface {
	Start(ctx context.Context, applicationContext ApplicationContext)
	Stop() error
}
