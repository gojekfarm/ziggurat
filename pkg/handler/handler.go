package handler

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"time"
)

var MessageHandler = func(app z.App, handlerFunc z.HandlerFunc) func(event basic.MessageEvent) {
	return func(event basic.MessageEvent) {
		metricTags := map[string]string{
			"topic_entity": event.TopicEntity,
			"kafka_topic":  event.Topic,
		}
		funcExecStartTime := time.Now()
		status := handlerFunc(event, app)
		funcExecEndTime := time.Now()
		app.MetricPublisher().Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case z.ProcessingSuccess:
			app.MetricPublisher().IncCounter("message_processing_success", 1, metricTags)
			logger.LogInfo("message handler: successfully processed message", map[string]interface{}{"msg": string(event.MessageValueBytes)})
		case z.SkipMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			logger.LogInfo("message handler: skipping message", nil)
		case z.RetryMessage:
			logger.LogInfo("message handler: retrying message", map[string]interface{}{"msg": string(event.MessageValueBytes)})
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			retryErr := app.MessageRetry().Retry(app, event)
			if retryErr != nil {
				panic(retryErr)
			}
		default:
			logger.LogError(fmt.Errorf("invalid handler return code got %d", status), "", nil)
		}
	}
}
