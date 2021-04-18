//+build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gojekfarm/ziggurat"
)

type FakeStreams struct{}
type FakeEvent struct {
	v []byte
}

func (f FakeEvent) Key() []byte {
	return nil
}

func (f FakeEvent) Value() []byte {
	return f.v
}

func (f FakeEvent) Headers() map[string]string {
	return map[string]string{}
}

func (f *FakeStreams) Stream(ctx context.Context, handler ziggurat.Handler) error {
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
				handler.Handle(ctx, FakeEvent{
					v: []byte(fmt.Sprintf("<<< Fake streamer message >>> [%d]", i)),
				})
				i++
				time.Sleep(2 * time.Second)
			}
		}
	}()
	return <-errChan
}

func FakeMiddleware(h ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		fmt.Println("[FAKE MIDDLEWARE]: ", time.Now())
		return h.Handle(ctx, event)
	})
}

func main() {
	fs := FakeStreams{}
	z := ziggurat.Ziggurat{}
	hf := ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		fmt.Println("Received message : => ", string(event.Value()))
		return nil
	})
	handler := FakeMiddleware(hf)
	if err := z.Run(context.Background(), &fs, handler); err != nil {
		fmt.Println("error: ", err)
	}
}
