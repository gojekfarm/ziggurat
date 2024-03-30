package ziggurat

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"github.com/stretchr/testify/mock"
	"sync/atomic"
	"testing"
	"time"
)

type MockConsumer struct {
	mock.Mock
	PollInterval time.Duration
}

func (m *MockConsumer) Consume(ctx context.Context, handler Handler) error {
	args := m.Called(ctx, handler)
	if m.PollInterval == 0 {
		m.PollInterval = 200 * time.Millisecond
	}

	keepAlive := true

	for keepAlive {
		select {
		case <-ctx.Done():
			keepAlive = false
		default:
			time.Sleep(m.PollInterval)
			handler.Handle(ctx, &Event{})
		}
	}
	return args.Error(0)

}

func TestZiggurat_Run(t *testing.T) {
	t.Run("clean shutdown error", func(t *testing.T) {
		var zig Ziggurat
		var msgCount int32
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		mc1 := MockConsumer{}
		mc2 := MockConsumer{}
		mc3 := MockConsumer{}
		handler := HandlerFunc(func(ctx context.Context, event *Event) {
			atomic.AddInt32(&msgCount, 1)
		})

		mc1.On("Consume", mock.Anything, mock.Anything).Return(nil)
		mc2.On("Consume", mock.Anything, mock.Anything).Return(nil)
		mc3.On("Consume", mock.Anything, mock.Anything).Return(nil)

		err := zig.Run(ctx, handler, &mc1, &mc2, &mc3)
		if !errors.Is(err, ErrCleanShutdown) {
			t.Errorf("expected %s got %s", ErrCleanShutdown.Error(), err.Error())
			return
		}
		if atomic.LoadInt32(&msgCount) < 1 {
			t.Error("handler not invoked")
		}
		t.Logf("message count:%d", msgCount)
	})

	t.Run("one of the consumers error out", func(t *testing.T) {
		var errorHandlerCalled atomic.Bool
		var zig Ziggurat
		zig.ErrorHandler = func(err error) {
			t.Logf("error handler invoked with error:%s", err.Error())
			errorHandlerCalled.Store(true)
		}
		var msgCount int32
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		mc1 := MockConsumer{}
		mc2 := MockConsumer{}
		mc3 := MockConsumer{}
		handler := HandlerFunc(func(ctx context.Context, event *Event) {
			atomic.AddInt32(&msgCount, 1)
		})

		mc1.On("Consume", mock.Anything, mock.Anything).Return(nil)
		mc2.On("Consume", mock.Anything, mock.Anything).Return(nil)
		mc3.On("Consume", mock.Anything, mock.Anything).Return(errors.New("mc3 errored out"))

		err := zig.Run(ctx, handler, &mc1, &mc2, &mc3)
		if err == nil {
			t.Error("expected error got nil")
			return
		}

		if !errorHandlerCalled.Load() {
			t.Error("error handler was never called")
		}

		if atomic.LoadInt32(&msgCount) < 1 {
			t.Error("handler not invoked")
		}
		t.Logf("message count:%d", msgCount)
	})

	t.Run("test shutdown timeout", func(t *testing.T) {
		var zig Ziggurat
		zig.ShutdownTimeout = 250 * time.Millisecond
		zig.Logger = logger.NewLogger(logger.LevelInfo)

		errCount := 1
		zig.ErrorHandler = func(err error) {
			errCount++
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()
		mc1 := MockConsumer{PollInterval: 10000 * time.Second}
		mc2 := MockConsumer{}
		mc3 := MockConsumer{}
		handler := HandlerFunc(func(ctx context.Context, event *Event) {})

		mc1.On("Consume", mock.Anything, mock.Anything).Return(nil)
		mc2.On("Consume", mock.Anything, mock.Anything).Return(errors.New("mc2 errored out"))
		mc3.On("Consume", mock.Anything, mock.Anything).Return(nil)

		err := zig.Run(ctx, handler, &mc1, &mc2, &mc3)
		if err == nil {
			t.Error("expected error got nil")
			return
		}

		if err.Error() != "shutdown timeout" {
			t.Errorf("expected error:%s got %s\n", "shutdown timeout", err.Error())
			return
		}

		if errCount < 1 {
			t.Errorf("expected an error count of 1 got %d\n", errCount)
		}

	})

}
