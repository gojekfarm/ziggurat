package ziggurat

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"sync"
	"time"
)

var ErrCleanShutdown = errors.New("clean shutdown of streams")

// Ziggurat serves as a container for message consumers to run in
// can be used without initialization
// var z ziggurat.Ziggurat
// z.run(ctx context.Context,s ziggurat.MessageConsumer,h ziggurat.Handler)
type Ziggurat struct {
	handler         Handler
	Logger          StructuredLogger
	ShutdownTimeout time.Duration
	ErrorHandler    func(err error)
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

	timeout := make(chan bool, 1)
	go func() {
		<-ctx.Done()
		<-time.After(z.ShutdownTimeout)
		z.Logger.Info("ziggurat consumer orchestration wait timeout")
		timeout <- true
		close(errChan)
	}()

	go func() {
		wg.Wait()
		close(errChan)
		close(timeout)
	}()

	var allErrs []error
	for consErr := range errChan {
		if z.ErrorHandler != nil {
			z.ErrorHandler(consErr)
		}
		allErrs = append(allErrs, consErr)
	}

	if <-timeout {
		return errors.New("shutdown timeout")
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
	if z.ShutdownTimeout == 0 {
		z.ShutdownTimeout = 6000 * time.Millisecond
	}
	if len(consumers) < 1 {
		panic("error: at least one ziggurat.MessageConsumer implementation should be provided")
	}

	if handler == nil {
		panic("error: handler cannot be nil")
	}
}
