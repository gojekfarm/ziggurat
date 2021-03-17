package statsd

import (
	"context"
	"runtime"
	"time"
)

func GoRoutinePublisher(ctx context.Context, interval time.Duration, s *Client) {
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C
	for {
		select {
		case <-done:
			t.Stop()
			return
		case <-tickerChan:
			s.client.Gauge("go_routine_count", int64(runtime.NumGoroutine()), 1.0)
		}
	}
}
