package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat"
)

type (
	Opts                   func(p *ProcessingStatusLogger)
	ProcessingStatusLogger struct {
		l       ziggurat.LeveledLogger
		handler ziggurat.Handler
	}
)

func (p *ProcessingStatusLogger) HandleMessage(event *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
	return p.LogStatus(p.handler).HandleMessage(event, ctx)
}

func WithLogger(logger ziggurat.LeveledLogger) func(p *ProcessingStatusLogger) {
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
		switch status {
		case ziggurat.ProcessingSuccess:
			p.l.Infof("successfully processed message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.RouteName, topic, messageEvent.Value)
		case ziggurat.RetryMessage:
			p.l.Infof("retrying message message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.RouteName, topic, messageEvent.Value)
		case ziggurat.SkipMessage:
			p.l.Infof("retrying message message ROUTE=%s TOPIC=%s VALUE=%s", messageEvent.RouteName, topic, messageEvent.Value)
		}
		return status
	})
}
