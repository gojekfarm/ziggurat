package ziggurat

import (
	"context"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat/logger"
)

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

	streams := MockKStreams{ConsumeFunc: func(ctx context.Context, handler Handler) chan error {
		done := make(chan error)
		go func() {
			<-ctx.Done()
			done <- nil
		}()
		return done
	}}

	z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event Event) error { return nil }))

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

	streams := MockKStreams{ConsumeFunc: func(ctx context.Context, handler Handler) chan error {
		done := make(chan error)
		go func() {
			<-ctx.Done()
			done <- nil
		}()
		return done
	}}
	z.streams = streams
	err := z.Run(ctx, streams, HandlerFunc(func(ctx context.Context, event Event) error { return nil }))
	if err != nil {
		t.Errorf("expected error to be nil")
	}
}
