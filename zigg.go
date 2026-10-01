package ziggurat

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gojekfarm/ziggurat/v2/logger"
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
	// buffered so a consumer's error send always succeeds immediately,
	// regardless of whether/when anything is still reading from errChan.
	errChan := make(chan error, len(consumers))
	for i := range consumers {
		go func(i int) {
			defer wg.Done()
			err := consumers[i].Consume(ctx, handler)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone) // the ONLY close in this function that isn't already guarded by allDone below
	}()

	select {
	case <-allDone:
		// every consumer finished on its own — fall through to collect errors below
	case <-ctx.Done():
		select {
		case <-allDone:
			// consumers happened to finish right as ctx was cancelled — treat as clean
		case <-time.After(z.ShutdownTimeout):
			z.Logger.Info("ziggurat consumer orchestration wait timeout")
			return errors.New("shutdown timeout")
		}
	}

	// Safe to close here: every consumer goroutine has already called wg.Done() by this
	// point (allDone is only closed after wg.Wait() returns), and each one only ever sends
	// to errChan *before* calling wg.Done() (see defer above), so nothing can still be
	// sending to errChan once we reach this line.
	close(errChan)

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
