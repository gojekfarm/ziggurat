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

	run := true
	for run {
		select {
		case <-done:
			t.Stop()
			s.logger.Error("stopping go-routine publisher", ctx.Err())
			run = false
		case <-tickerChan:
			s.logger.Error(
				"error publishing go-routine count",
				s.Gauge("num_goroutine", int64(runtime.NumGoroutine()), nil))
		}
	}
}
