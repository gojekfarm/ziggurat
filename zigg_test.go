package ziggurat

import (
	"context"
	"testing"
	"time"
)

func TestZigguratStartStop(t *testing.T) {
	isStartCalled := false
	isStopCalled := false
	ctx, cancelFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelFunc()
	z := NewApp(WithContext(ctx), WithLogLevel("disabled"))
	z.OnStart(func(a App) {
		isStartCalled = true
	})
	z.OnStop(func(a App) {
		isStopCalled = true
	})
	kstreams := NewKafkaStreams()
	kstreams.StartFunc = func(a App) (chan struct{}, error) {
		done := make(chan struct{})
		ctxDone := a.Context().Done()
		go func() {
			<-ctxDone
			close(done)
		}()
		return done, nil
	}

	z.streams = kstreams

	<-z.Run(HandlerFunc(func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	}),
		Routes{
			"foo": {},
		})
	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZigguratRun(t *testing.T) {
	z := NewApp(WithLogLevel("disabled"))
	handler := HandlerFunc(func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})
	z.OnStart(func(a App) {
		if !z.IsRunning() {
			t.Errorf("expected app to be running")
		}
	})
	<-z.Run(handler, Routes{"foo": {}})
}
