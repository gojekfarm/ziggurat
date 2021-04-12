package ziggurat

import (
	"testing"
	"time"
)

func TestWaitAll(t *testing.T) {
	var chans []chan struct{}

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
		time.Sleep(250 * time.Millisecond)
		t.Errorf("wait time exceeded, expected to completed by %dms", 200)
	}()

	WaitAll(chans...)
}
