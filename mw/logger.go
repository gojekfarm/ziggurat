package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat"
)

type (
	Opts                   func(p *ProcessingStatusLogger)
	ProcessingStatusLogger struct {
		l       ziggurat.StructuredLogger
		handler ziggurat.Handler
	}
)

func (p *ProcessingStatusLogger) HandleMessage(event *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
	return p.LogStatus(p.handler).HandleMessage(event, ctx)
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
		p.l = ziggurat.NewLogger("info")
	}
	return p
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(messageEvent *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		status := next.HandleMessage(messageEvent, ctx)
		topic := messageEvent.Attribute("kafka-topic")
		args := map[string]interface{}{"ROUTE": messageEvent.RouteName, "TOPIC": topic, "VALUE": messageEvent.Value}
		switch status {
		case ziggurat.ProcessingSuccess:
			p.l.Info("successfully processed message", args)
		case ziggurat.RetryMessage:
			p.l.Info("retrying message message", args)
		case ziggurat.SkipMessage:
			p.l.Info("skipping message", args)
		}
		return status
	})
}
