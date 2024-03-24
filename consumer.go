package ziggurat

import "context"

type MessageConsumer interface {
	Consume(ctx context.Context, handler Handler) error
}
