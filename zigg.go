package ziggurat

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/logger"
	"sync"
	"time"
)

const defaultWaitTimeout = 5000 * time.Millisecond

var ErrCleanShutdown = errors.New("clean shutdown of streams")

// Ziggurat serves as a container for streams to run in
// can be used without initialization
// var z ziggurat.Ziggurat
// z.run(ctx context.Context,s ziggurat.MessageConsumer,h ziggurat.Handler)
type Ziggurat struct {
	handler      Handler
	Logger       StructuredLogger
	WaitTimeout  time.Duration
	ErrorHandler func(err error)
}

func (z *Ziggurat) Run(ctx context.Context, handler Handler, consumers ...MessageConsumer) error {

	z.mustInit(consumers, handler)

	var wg sync.WaitGroup
	wg.Add(len(consumers))
	errChan := make(chan error)
	for i := range consumers {
		go func(i int) {
			err := consumers[i].Consume(ctx, handler)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}(i)
	}

	go func() {
		if timeout := waitWithTimeout(&wg, z.WaitTimeout); timeout {
			z.Logger.Warn("wait timed out")
		}
		close(errChan)
	}()

	var allErrs []error
	for consErr := range errChan {
		if z.ErrorHandler != nil {
			z.ErrorHandler(consErr)
		}
		allErrs = append(allErrs, consErr)
	}

	if len(allErrs) > 0 {
		return errors.Join(allErrs...)
	}

	return ErrCleanShutdown

}

func (z *Ziggurat) mustInit(consumers []MessageConsumer, handler Handler) {
	if z.Logger == nil {
		z.Logger = logger.NOOP
	}
	if len(consumers) < 1 {
		panic("error: at least one ziggurat.MessageConsumer implementation should be provided")
	}

	if handler == nil {
		panic("error: handler cannot be nil")
	}
	if z.WaitTimeout == 0 {
		z.WaitTimeout = defaultWaitTimeout
	}

}

func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timeoutAfter := time.After(timeout)
	select {
	case <-done:
		return false
	case <-timeoutAfter:
		return true
	}
}
