package statsd

import (
	"context"
	"runtime"
	"time"
)

var publishGoRoutines = func(ctx context.Context, interval time.Duration, s *Client) {
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C

	for {
		select {
		case <-done:
			t.Stop()
			s.logger.Error("stopping go-routine publisher", ctx.Err())
		case <-tickerChan:
			s.logger.Error(
				"error publishing go-routine count",
				s.Gauge("num_goroutine", int64(runtime.NumGoroutine()), nil))
		}
	}
}
