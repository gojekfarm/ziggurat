package zmw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"time"
)

var Terminal = func(next ztype.MessageHandler) ztype.MessageHandler {
	return ztype.HandlerFunc(func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		metricTags := map[string]string{
			"topic_entity": event.StreamRoute,
		}
		funcExecStartTime := time.Now()
		status := next.HandleMessage(event, app)
		funcExecEndTime := time.Now()
		app.MetricPublisher().Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case ztype.ProcessingSuccess:
			app.MetricPublisher().IncCounter("message_processing_success", 1, metricTags)
		case ztype.SkipMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
		case ztype.RetryMessage:
			app.MetricPublisher().IncCounter("message_processing_failure_skip", 1, metricTags)
			retryErr := app.MessageRetry().Retry(app, event)
			if retryErr != nil {
				panic(retryErr)
			}
		default:
			zlog.LogError(fmt.Errorf("invalid handler return code got %d", status), "", nil)
		}
		return status
	})
}
