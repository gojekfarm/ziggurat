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
			logInfo("message handler: successfully processed message", map[string]interface{}{"msg": string(event.MessageValueBytes)})
		case SkipMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			logInfo("message handler: skipping message", nil)
		case RetryMessage:
			logInfo("message handler: retrying message", map[string]interface{}{"msg": string(event.MessageValueBytes)})
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			retryErr := app.MessageRetry().Retry(app, event)
			if retryErr != nil {
				panic(retryErr)
			}
		default:
			logError(fmt.Errorf("invalid handler return code got %d", status), "", nil)
		}
	}
}
