package proclog

import (
	"context"

	"github.com/gojekfarm/ziggurat"
)

type (
	ProcLogger struct {
		Logger ziggurat.StructuredLogger
		Next   ziggurat.Handler
	}
)

func (p *ProcLogger) Handle(ctx context.Context, event ziggurat.Event) error {
	if p.Next == nil {
		panic("[process status logger] handler cannot be nil")
	}
	return p.LogStatus(p.Next).Handle(ctx, event)
}

func (p *ProcLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, messageEvent ziggurat.Event) error {
		if p.Logger == nil {
			return next.Handle(ctx, messageEvent)
		}

		err := next.Handle(ctx, messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute]}
		if err != nil {
			p.Logger.Error("message processing failed", err, args)
		} else {
			p.Logger.Info("message processing succeeded", args)
		}
		return err
	}
	return ziggurat.HandlerFunc(f)
}
