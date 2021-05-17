package statsd

import (
	"context"
	"errors"
	"runtime"
	"runtime/metrics"
	"time"
)

func goRoutinePublisher(ctx context.Context, interval time.Duration, s *Client) {
	done := ctx.Done()
	t := time.NewTicker(interval)
	tickerChan := t.C

	const metricName = "/sched/goroutines:goroutines"
	sample := make([]metrics.Sample, 1)
	sample[0].Name = metricName

	for {
		select {
		case <-done:
			t.Stop()
			return
		case <-tickerChan:
			metrics.Read(sample)
			if sample[0].Value.Kind() == metrics.KindBad {
				s.logger.Error("error publishing metric", errors.New("bad metric kind"))
			} else {
				s.logger.Error("error publishing goroutine count",
					s.Gauge("num_alive_goroutine",
						int64(sample[0].Value.Uint64()),
						nil),
				)
			}
			s.logger.Error("error publishing go-routine count",
				s.Gauge("num_goroutine",
					int64(runtime.NumGoroutine()),
					nil),
			)
		}
	}
}
