package mw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

func ProcessingStatusLogger(next ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event, app ziggurat.AppContext) ziggurat.ProcessStatus {
		args := map[string]interface{}{
			"ROUTE": messageEvent.RouteName(),
			"VALUE": fmt.Sprintf("%s", messageEvent.MessageValue()),
		}
		status := next.HandleMessage(messageEvent, app)
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
