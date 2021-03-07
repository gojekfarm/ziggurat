package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat"
)

type (
	ProcessingStatusLogger struct {
		handler ziggurat.Handler
		logger  ziggurat.StructuredLogger
	}
)

type Opts func(psl *ProcessingStatusLogger)

func WithHandler(h ziggurat.Handler) Opts {
	return func(psl *ProcessingStatusLogger) {
		psl.handler = h
	}
}

func WithLogger(l ziggurat.StructuredLogger) Opts {
	return func(psl *ProcessingStatusLogger) {
		psl.logger = l
	}
}

func NewProcessStatusLogger(opts ...Opts) *ProcessingStatusLogger {
	psl := &ProcessingStatusLogger{}
	for _, opt := range opts {
		opt(psl)
	}

	return psl
}

func (p *ProcessingStatusLogger) HandleEvent(ctx context.Context, event ziggurat.Event) error {
	if p.handler == nil {
		panic("handler cannot be nil")
	}
	return p.LogStatus(p.handler).HandleEvent(ctx, event)
}

func (p *ProcessingStatusLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
		if p.logger == nil {
			return next.HandleEvent(ctx, messageEvent)
		}
		err := next.HandleEvent(ctx, messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute], "value": messageEvent.Value()}
		if err != nil {
			p.logger.Error("message processing failed", err, args)
		} else {
			p.logger.Info("message processing succeeded", args)
		}
		return err
	})
}
