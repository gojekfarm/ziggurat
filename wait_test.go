package ziggurat

import (
	"context"
	"testing"
	"time"
)

func TestWaitAll(t *testing.T) {
	var chans []chan struct{}
	ctx, cfn := context.WithCancel(context.Background())

	for i := 0; i < 5; i++ {
		chans = append(chans, make(chan struct{}))
	}
	for _, c := range chans {
		cCopy := c
		go func() {
			time.Sleep(200 * time.Millisecond)
			cCopy <- struct{}{}
		}()
	}

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(250 * time.Millisecond):
			t.Error("expected to complete in 200ms but completed in 250ms")

		}
	}()

	WaitAll(chans...)
	cfn()

}
