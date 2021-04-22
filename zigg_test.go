package ziggurat

import (
	"context"
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
	z := &Ziggurat{Logger: logger.NewJSONLogger("disabled")}
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

	z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) error { return nil }))

	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZiggurat_Run(t *testing.T) {
	z := &Ziggurat{Logger: logger.NewJSONLogger("disabled")}
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()

	streams := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return nil
	}}
	z.streams = streams
	err := z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) error { return nil }))
	if err != nil {
		t.Errorf("expected error to be nil")
	}
}

func TestZiggurat_RunAll(t *testing.T) {
	var z Ziggurat
	z.Logger = logger.DiscardLogger{}
	c, cfn := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cfn()
	ms := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	msTwo := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return ctx.Err()
	}}

	err := z.RunAll(c, HandlerFunc(func(ctx context.Context, event *Event) error {
		return nil
	}), ms, msTwo)

	runAllErr := err.(ErrRunAll)

	for _, e := range runAllErr.Inspect() {
		if e != context.DeadlineExceeded {
			t.Errorf("expected %s but got %s", context.DeadlineExceeded.Error(), e.Error())
		}
	}
}

func TestErrRunAll_HandlerExec(t *testing.T) {
	var z Ziggurat
	callCount := 0
	streamOne := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				event := Event{EventType: "foo"}
				handler.Handle(ctx, &event)
			}
		}
	}}
	streamTwo := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				event := Event{EventType: "bar"}
				handler.Handle(ctx, &event)
			}
		}
	}}
	h := HandlerFunc(func(ctx context.Context, event *Event) error {
		if event.EventType == "foo" || event.EventType == "bar" {
			callCount++
		}
		return nil
	})
	c, cfn := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cfn()
	z.RunAll(c, h, streamOne, streamTwo)
	if callCount < 1 {
		t.Errorf("expected call count to be greater than 1 got %d", callCount)
	}
}
