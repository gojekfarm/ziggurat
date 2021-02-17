package mw

import (
	"github.com/gojekfarm/ziggurat"
)

type (
	ProcessingStatusLogger struct {
		Logger  ziggurat.StructuredLogger
		Handler ziggurat.Handler
	}
)

func (p *ProcessingStatusLogger) HandleEvent(event ziggurat.Event) error {
	return p.LogStatus(p.Handler).HandleEvent(event)
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) error {
		if p.Logger == nil {
			return next.HandleEvent(messageEvent)
		}
		err := next.HandleEvent(messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		if err != nil {
			p.Logger.Error("message processing failed", err, args)
		} else {
			p.Logger.Info("message processing succeeded", args)
		}
		return err
	})
}
