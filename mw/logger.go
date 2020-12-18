package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat"
)

type (
	Opts                   func(p *ProcessingStatusLogger)
	ProcessingStatusLogger struct {
		l       ziggurat.LeveledLogger
		handler ziggurat.MessageHandler
	}
)

func (p *ProcessingStatusLogger) HandleMessage(event ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
	return p.LogStatus(p.handler).HandleMessage(event, ctx)
}

func WithLogger(logger ziggurat.LeveledLogger) func(p *ProcessingStatusLogger) {
	return func(p *ProcessingStatusLogger) {
		p.l = logger
	}
}

func WithHandler(h ziggurat.MessageHandler) func(p *ProcessingStatusLogger) {
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

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
		status := next.HandleMessage(messageEvent, ctx)
		switch status {
		case ziggurat.ProcessingSuccess:
			p.l.Infof("successfully processed message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.StreamRoute, messageEvent.Topic, messageEvent.MessageValue)
		case ziggurat.RetryMessage:
			p.l.Infof("retrying message message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.StreamRoute, messageEvent.Topic, messageEvent.MessageValue)
		case ziggurat.SkipMessage:
			p.l.Infof("retrying message message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.StreamRoute, messageEvent.Topic, messageEvent.MessageValue)
		}
		return status
	})
}
