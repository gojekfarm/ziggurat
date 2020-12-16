package zigg

import (
	"context"
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"testing"
	"time"
)

func TestZigguratStartStop(t *testing.T) {
	isStartCalled := false
	isStopCalled := false
	ctx, cancelFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelFunc()
	z := New(WithContext(ctx), WithLogLevel("disabled"))
	z.OnStart(func(a ztype.App) {
		isStartCalled = true
	})
	z.OnStop(func(a ztype.App) {
		isStopCalled = true
	})
	kstreams := mock.NewKafkaStreams()
	kstreams.StartFunc = func(a ztype.App) (chan struct{}, error) {
		done := make(chan struct{})
		ctxDone := a.Context().Done()
		go func() {
			<-ctxDone
			close(done)
		}()
		return done, nil
	}

	z.streams = kstreams

	<-z.Run(ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	}),
		zbase.Routes{
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
	z := New(WithLogLevel("disabled"))
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	z.OnStart(func(a ztype.App) {
		if !z.IsRunning() {
			t.Errorf("expected app to be running")
		}
	})
	<-z.Run(handler, zbase.Routes{"foo": {}})
}
