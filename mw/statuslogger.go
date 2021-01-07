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

func (p *ProcessingStatusLogger) HandleEvent(event ziggurat.Event) ziggurat.ProcessStatus {
	return p.LogStatus(p.Handler).HandleEvent(event)
}

func WithLogger(logger ziggurat.StructuredLogger) func(p *ProcessingStatusLogger) {
	return func(p *ProcessingStatusLogger) {
		p.Logger = logger
	}
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		status := next.HandleEvent(messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		switch status {
		case ziggurat.ProcessingSuccess:
			args["status"] = "success"
			p.Logger.Info("", args)
			break
		case ziggurat.RetryMessage:
			args["status"] = "retry"
			p.Logger.Info("", args)
			break
		case ziggurat.SkipMessage:
			args["status"] = "skip"
			p.Logger.Info("", args)
			break
		}
		return status
	})
}
