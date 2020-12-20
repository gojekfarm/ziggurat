package ziggurat

import (
	"context"
	"testing"
	"time"
)

func TestZigguratStartStop(t *testing.T) {
	isStartCalled := false
	isStopCalled := false
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()
	z := &Ziggurat{Logger: NewLogger("disabled")}
	z.StartFunc(func(ctx context.Context) {
		isStartCalled = true
	})
	z.StopFunc(func() {
		isStopCalled = true
	})

	z.streams = MockKStreams{ConsumeFunc: func(ctx context.Context, routes Routes, handler Handler) chan error {
		done := make(chan error)
		go func() {
			<-ctx.Done()
			done <- nil
		}()
		return done
	}}

	<-z.Run(ctx, HandlerFunc(func(messageEvent Message, ctx context.Context) ProcessStatus { return ProcessingSuccess }), Routes{"foo": {}})

	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZigguratRun(t *testing.T) {
	z := &Ziggurat{}
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()
	z.StartFunc(func(ctx context.Context) {
		if !z.IsRunning() {
			t.Errorf("expected app to be running state")
		}
	})
	streams := MockKStreams{ConsumeFunc: func(ctx context.Context, routes Routes, handler Handler) chan error {
		done := make(chan error)
		go func() {
			<-ctx.Done()
			done <- nil
		}()
		return done
	}}
	z.streams = streams
	<-z.Run(ctx, HandlerFunc(func(messageEvent Message, ctx context.Context) ProcessStatus { return ProcessingSuccess }), Routes{"foo": {}})
}
