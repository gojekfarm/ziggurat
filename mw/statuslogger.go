package mw

import (
	"context"

	"github.com/gojekfarm/ziggurat"
)

type (
	ProcessingStatusLogger struct {
		Logger  ziggurat.StructuredLogger
		Handler ziggurat.Handler
	}
)

func (p *ProcessingStatusLogger) Handle(ctx context.Context, event ziggurat.Event) error {
	if p.Handler == nil {
		panic("[process status logger] handler cannot be nil")
	}
	return p.LogStatus(p.Handler).Handle(ctx, event)
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
		if p.Logger == nil {
			return next.Handle(ctx, messageEvent)
		}

		err := next.Handle(ctx, messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		if err != nil {
			p.Logger.Error("message processing failed", err, args)
		} else {
			p.Logger.Info("message processing succeeded", args)
		}
		return err
	})
}
