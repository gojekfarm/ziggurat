package metrics

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"runtime"
	"time"
)

func GoRoutinePublisher(ctx context.Context, interval time.Duration, s *Client) {
	ziggurat.LogInfo("statsd: starting go-routine publisher", nil)
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C
	for {
		select {
		case <-done:
			t.Stop()
			ziggurat.LogInfo("statsd: halting go-routine publisher", nil)
			return
		case <-tickerChan:
			s.client.Gauge("go_routine_count", int64(runtime.NumGoroutine()), 1.0)
		}
	}
}
