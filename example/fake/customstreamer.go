package main

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"time"
)

type FakeStreams struct{}
type FakeEvent struct {
	c context.Context
	v []byte
}

func (f FakeEvent) Value() []byte {
	return f.v
}

func (f FakeEvent) Headers() map[string]string {
	return map[string]string{}
}

func (f FakeEvent) Context() context.Context {
	return f.c
}

func (f *FakeStreams) Stream(ctx context.Context, handler ziggurat.Handler) chan error {
	errChan := make(chan error)
	done := ctx.Done()
	i := 0
	go func() {
		for {
			select {
			case <-done:
				errChan <- ctx.Err()
				return
			default:
				handler.HandleEvent(FakeEvent{
					c: ctx,
					v: []byte(fmt.Sprintf("<<< Fake streamer message >>> [%d]", i)),
				})
				i++
				time.Sleep(2 * time.Second)
			}
		}
	}()
	return errChan
}

func main() {
	fs := &FakeStreams{}
	z := &ziggurat.Ziggurat{}
	<-z.Run(context.Background(), fs, ziggurat.HandlerFunc(func(event ziggurat.Event) ziggurat.ProcessStatus {
		fmt.Println("Received message : => ", string(event.Value()))
		return ziggurat.ProcessingSuccess
	}))
}
