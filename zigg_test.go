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

	z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) interface{} { return nil }))

	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZigguratRun(t *testing.T) {
	z := &Ziggurat{Logger: logger.NewJSONLogger("disabled")}
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()

	streams := mockStreams{ConsumeFunc: func(ctx context.Context, handler Handler) error {
		<-ctx.Done()
		return nil
	}}
	z.streams = streams
	err := z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event *Event) interface{} { return nil }))
	if err != nil {
		t.Errorf("expected error to be nil")
	}
}
