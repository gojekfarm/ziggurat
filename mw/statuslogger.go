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

func (p *ProcessingStatusLogger) HandleEvent(event ziggurat.Event, ctx context.Context) error {
	return p.LogStatus(p.Handler).HandleEvent(event, ctx)
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event, ctx context.Context) error {
		if p.Logger == nil {
			return next.HandleEvent(messageEvent, ctx)
		}
		err := next.HandleEvent(messageEvent, ctx)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		if err != nil {
			p.Logger.Error("message processing failed", err, args)
		} else {
			p.Logger.Info("message processing succeeded", args)
		}
		return err
	})
}
