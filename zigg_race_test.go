package ziggurat

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

type raceMockConsumer struct{ exitAfter time.Duration }

func (c *raceMockConsumer) Consume(ctx context.Context, h Handler) error {
	<-ctx.Done()
	// jitter the exit time so across many runs it lands on both sides of ShutdownTimeout
	jitter := time.Duration(rand.Int63n(int64(20*time.Millisecond))) - 10*time.Millisecond
	time.Sleep(c.exitAfter + jitter)
	return ctx.Err()
}

func TestRun_DoubleCloseRace(t *testing.T) {
	for i := 0; i < 3000; i++ {
		z := &Ziggurat{ShutdownTimeout: 50 * time.Millisecond}
		ctx, cancel := context.WithCancel(context.Background())
		consumer := &raceMockConsumer{exitAfter: 50 * time.Millisecond} // right at the ShutdownTimeout boundary
		h := HandlerFunc(func(ctx context.Context, e *Event) {})

		go func() {
			time.Sleep(2 * time.Millisecond)
			cancel()
		}()

		_ = z.Run(ctx, h, consumer)
	}
}
