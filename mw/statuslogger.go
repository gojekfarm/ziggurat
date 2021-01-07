package mw

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

type (
	Opts                   func(p *ProcessingStatusLogger)
	ProcessingStatusLogger struct {
		l       ziggurat.StructuredLogger
		handler ziggurat.Handler
	}
)

func (p *ProcessingStatusLogger) HandleEvent(event ziggurat.Message) ziggurat.ProcessStatus {
	return p.LogStatus(p.handler).HandleEvent(event)
}

func WithLogger(logger ziggurat.StructuredLogger) func(p *ProcessingStatusLogger) {
	return func(p *ProcessingStatusLogger) {
		p.l = logger
	}
}

func WithHandler(h ziggurat.Handler) func(p *ProcessingStatusLogger) {
	return func(p *ProcessingStatusLogger) {
		p.handler = h
	}
}

func NewProcessingStatusLogger(opts ...Opts) *ProcessingStatusLogger {
	p := &ProcessingStatusLogger{}
	for _, opt := range opts {
		opt(p)
	}
	if p.l == nil {
		p.l = logger.NewJSONLogger("info")
	}
	return p
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
		status := next.HandleEvent(messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		switch status {
		case ziggurat.ProcessingSuccess:
			args["status"] = "success"
			p.l.Info("", args)
			break
		case ziggurat.RetryMessage:
			args["status"] = "retry"
			p.l.Info("", args)
			break
		case ziggurat.SkipMessage:
			args["status"] = "skip"
			p.l.Info("", args)
			break
		}
		return status
	})
}
