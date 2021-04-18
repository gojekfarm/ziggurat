package router

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat"
)

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := New()
	called := false
	dr.HandleFunc("foo", func(ctx context.Context, event *ziggurat.Event) error {
		called = true
		return nil
	})

	dr.HandleFunc("bar", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})

	dr.Handle(context.Background(), &ziggurat.Event{Path: "foo"})

	if !called {
		t.Errorf("expected foo handler to be called")
	}

}

func TestDefaultRouter_NotFoundHandler(t *testing.T) {
	notFoundHandlerCalled := false
	dr := New(WithNotFoundHandler(func(ctx context.Context, event *ziggurat.Event) error {
		notFoundHandlerCalled = true
		return nil
	}))

	dr.HandleFunc("foo", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})

	dr.Handle(context.Background(), &ziggurat.Event{Path: "bar"})

	if !notFoundHandlerCalled {
		t.Errorf("expected %v got %v", true, notFoundHandlerCalled)
	}

}
