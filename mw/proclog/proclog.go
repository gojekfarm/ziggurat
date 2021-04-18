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
func (p *ProcLogger) Handle(ctx context.Context, event *ziggurat.Event) interface{} {
	if p.Next == nil {
		panic("[process status logger] handler cannot be nil")
	}
	return p.LogStatus(p.Next).Handle(ctx, event)
}

// LogStatus logs the processing status of the handler
func (p *ProcLogger) LogStatus(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, messageEvent ziggurat.Event) interface{} {
		if p.Logger == nil {
			return next.Handle(ctx, messageEvent)
		}
		retVal := next.Handle(ctx, messageEvent)
		args := map[string]interface{}{"route": messageEvent.Headers()[ziggurat.HeaderMessageRoute]}

		switch retVal.(type) {
		case error:
			args["status"] = "PROCESSING_FAILED"
			p.Logger.Error("process status logger", retVal.(error), args)
			return retVal
		default:
			args["status"] = "PROCESSING_SUCCEEDED"
			args["return-value"] = retVal
			p.Logger.Info("process status logger", args)
			return retVal
		}

	}
	return ziggurat.HandlerFunc(f)
}
