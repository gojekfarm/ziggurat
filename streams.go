package ziggurat

import "context"

type Streamer interface {
	Stream(ctx context.Context, handler Handler) error
}
