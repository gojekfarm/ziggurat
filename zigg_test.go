package ziggurat

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat/logger"
)

type mockStreams struct {
	ConsumeFunc func(ctx context.Context, handler Handler) error
}

func (m mockStreams) Stream(ctx context.Context, handler Handler) error {
	return m.ConsumeFunc(ctx, handler)
}

func TestZigguratStartStop(t *testing.T) {
	isStartCalled := false
	isStopCalled := false
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()
	z := &Ziggurat{Logger: logger.NOOP}
	z.StartFunc(func(ctx context.Context) {
		isStartCalled = true
	})
	z.StopFunc(func() {
		isStopCalled = true
	})

	streams := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return ctx.Err()
	}}

	_ = z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) error { return nil }))

	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZiggurat_Run(t *testing.T) {
	z := &Ziggurat{Logger: logger.NOOP}
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()

	streams := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return nil
	}}

	err := z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) error { return nil }))
	if err != nil {
		t.Errorf("expected error to be nil")
	}
}

func TestZiggurat_RunAll(t *testing.T) {
	var z Ziggurat
	streamOne := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	streamTwo := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	ctx, cfn := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cfn()
	err := z.RunAll(ctx, HandlerFunc(func(ctx context.Context, event *Event) error {
		return nil
	}), streamOne, streamTwo)

	if err != nil && !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Errorf("expected error to contain %s got %s", context.DeadlineExceeded.Error(), err.Error())
	}
}

func TestZiggurat_MultiStreamHandlerExec(t *testing.T) {
	var z Ziggurat
	var handlerCallCount int64
	streamOne := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		return handler.Handle(ctx, &Event{EventType: "foo"})
	}}

	streamTwo := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		return handler.Handle(ctx, &Event{EventType: "bar"})
	}}

	_ = z.RunAll(context.Background(), HandlerFunc(func(ctx context.Context, event *Event) error {
		if event.EventType == "foo" || event.EventType == "bar" {
			atomic.AddInt64(&handlerCallCount, 1)
		}
		return nil
	}), streamOne, streamTwo)

	if val := atomic.LoadInt64(&handlerCallCount); val < 2 {
		t.Errorf("expected handler call count to be %d got %d", 2, val)
	}

}
