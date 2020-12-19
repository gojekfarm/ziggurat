package ziggurat

import (
	"context"
	"testing"
)

func TestZigguratStartStop(t *testing.T) {
	isStartCalled := false
	isStopCalled := false
	z := NewApp(WithLogger(NewLogger("disabled")))
	z.StartFunc(func(ctx context.Context) {
		isStartCalled = true
	})
	z.StopFunc(func() {
		isStopCalled = true
	})

	z.streams = NewMockKafkaStreams()

	<-z.Run(context.Background(), HandlerFunc(func(messageEvent *Message, ctx context.Context) ProcessStatus { return ProcessingSuccess }), Routes{"foo": {}})

	if !isStartCalled {
		t.Error("expected start callback to be called")
	}
	if !isStopCalled {
		t.Error("expected stop callback to be called")
	}
}

func TestZigguratRun(t *testing.T) {
	z := &Ziggurat{}
	z.StartFunc(func(ctx context.Context) {
		if !z.IsRunning() {
			t.Errorf("expected app to be running state")
		}
	})
	<-z.Run(context.Background(), HandlerFunc(func(messageEvent *Message, ctx context.Context) ProcessStatus { return ProcessingSuccess }), Routes{"foo": {}})
}
