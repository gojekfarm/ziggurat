package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
)

type Worker struct {
	concurrency int
	sendCh      chan *kafka.Message
	doneCh      chan struct{}
}

func NewWorker(concurrency int) *Worker {
	return &Worker{
		concurrency: concurrency,
		sendCh:      make(chan *kafka.Message),
		doneCh:      make(chan struct{}, concurrency),
	}
}

func (w *Worker) run(ctx context.Context, f func(*kafka.Message)) (chan *kafka.Message, chan struct{}) {
	wg := &sync.WaitGroup{}
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			done := ctx.Done()
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case msg, ok := <-w.sendCh:
					if !ok {
						return
					}
					f(msg)
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(w.doneCh)
	}()
	return w.sendCh, w.doneCh
}
