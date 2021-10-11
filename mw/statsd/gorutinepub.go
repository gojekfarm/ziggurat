package statsd

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func publishGoRoutines(ctx context.Context, interval time.Duration, s *Client) error {
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C

	for {
		select {
		case <-done:
			t.Stop()
			return fmt.Errorf("stopped goroutine publisher: %w", ctx.Err())
		case <-tickerChan:
			s.logger.Error("error publishing go-routine count", s.Gauge("num_goroutine", int64(runtime.NumGoroutine()), nil))
		}
	}
}
