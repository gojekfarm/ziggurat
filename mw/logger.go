package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat"
)

type ProcessingStatusLogger struct {
	l ziggurat.LeveledLogger
}

func NewProcessingStatusLogger(logger ziggurat.LeveledLogger) *ProcessingStatusLogger {
	p := &ProcessingStatusLogger{l: logger}
	if p.l == nil {
		p.l = ziggurat.NewLogger("info")
	}
	return p
}

func (p *ProcessingStatusLogger) Log(next ziggurat.MessageHandler) ziggurat.MessageHandler {
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
