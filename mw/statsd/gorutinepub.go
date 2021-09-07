package statsd

import (
	"context"
	"runtime"
	"time"
)

func goRoutinePublisher(ctx context.Context, interval time.Duration, s *Client) {
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C

	for {
		select {
		case <-done:
			t.Stop()
			return
		case <-tickerChan:
			s.logger.Error("error publishing go-routine count", s.Gauge("num_goroutine", int64(runtime.NumGoroutine()), nil))
		}
	}
}
