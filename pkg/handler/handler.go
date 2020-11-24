package handler

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"time"
)

var DefaultTerminalMW = func(next z.MessageHandler) z.MessageHandler {
	return z.HandlerFunc(func(event basic.MessageEvent, app z.App) z.ProcessStatus {
		metricTags := map[string]string{
			"topic_entity": event.StreamRoute,
			"kafka_topic":  event.Topic,
		}
		funcExecStartTime := time.Now()
		status := next.HandleMessage(event, app)
		funcExecEndTime := time.Now()
		app.MetricPublisher().Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case z.ProcessingSuccess:
			app.MetricPublisher().IncCounter("message_processing_success", 1, metricTags)
		case z.SkipMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
		case z.RetryMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			retryErr := app.MessageRetry().Retry(app, event)
			if retryErr != nil {
				panic(retryErr)
			}
		default:
			logger.LogError(fmt.Errorf("invalid handler return code got %d", status), "", nil)
		}
		return status
	})
}
