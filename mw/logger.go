package mw

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

func ProcessingStatusLogger(next ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
		args := map[string]interface{}{
			"ROUTE": messageEvent.StreamRoute,
			"VALUE": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		status := next.HandleMessage(messageEvent, ctx)
		switch status {
		case ziggurat.ProcessingSuccess:
			ziggurat.LogInfo("[Msg logger]: successfully processed message", args)
		case ziggurat.RetryMessage:
			ziggurat.LogInfo("[Msg logger]: retrying message", args)
		case ziggurat.SkipMessage:
			ziggurat.LogInfo("[Msg logger]: skipping message", args)
		}
		return status
	})
}
