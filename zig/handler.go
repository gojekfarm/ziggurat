package zig

import (
	"fmt"
	"time"
)

func MessageHandler(app App, handlerFunc HandlerFunc) func(event MessageEvent) {
	return func(event MessageEvent) {
		metricTags := map[string]string{
			"topic_entity": event.TopicEntity,
			"kafka_topic":  event.Topic,
		}
		funcExecStartTime := time.Now()
		status := handlerFunc(event, app)
		funcExecEndTime := time.Now()
		app.MetricPublisher().Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case ProcessingSuccess:
			app.MetricPublisher().IncCounter("message_processing_success", 1, metricTags)
			logInfo("ziggurat message handler: successfully processed message")
		case SkipMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			logInfo("ziggurat message handler: skipping message")
		case RetryMessage:
			logInfo("ziggurat message handler: retrying message")
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			retryErr := app.MessageRetry().Retry(app, event)
			panic(retryErr)
		default:
			logError(fmt.Errorf("invalid handler return code got %d", status), "")
		}
	}
}
