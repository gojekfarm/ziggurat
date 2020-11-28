package kstream

import (
	"context"
	"sync"
	"time"
)

type Worker struct {
	concurrency int
	workChan    chan func()
}

func NewWorker(concurrency int) *Worker {
	return &Worker{concurrency: concurrency, workChan: make(chan func())}
}

func (w *Worker) enqueue(ctx context.Context, f func()) {
	done := ctx.Done()
	for {
		select {
		case <-done:
			return
		case w.workChan <- f:
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *Worker) run(ctx context.Context) chan struct{} {
	finishedChan := make(chan struct{})
	doneChan := ctx.Done()
	wg := &sync.WaitGroup{}
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-doneChan:
					wg.Done()
					return
				case f, ok := <-w.workChan:
					if ok {
						f()
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(finishedChan)
	}()
	return finishedChan
}

func (w *Worker) close() {
	close(w.workChan)
}
