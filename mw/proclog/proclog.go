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

// Handle implements the ziggurat.Handler interface
func (p *ProcLogger) Handle(ctx context.Context, event ziggurat.Event) interface{} {
	if p.Next == nil {
		panic("[process status logger] handler cannot be nil")
	}
	return p.LogStatus(p.Next).Handle(ctx, event)
}

// LogStatus logs the processing status of the handler
func (p *ProcLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, messageEvent ziggurat.Event) error {
		if p.Logger == nil {
			return next.Handle(ctx, messageEvent)
		}

		err := next.Handle(ctx, messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute]}
		if err != nil {
			args["status"] = "SUCCESS"
			p.Logger.Error("proc logger", err, args)
		} else {
			args["status"] = "FAILED"
			p.Logger.Info("proc logger", args)
		}
		return err
	}
	return ziggurat.HandlerFunc(f)
}
