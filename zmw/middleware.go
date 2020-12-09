package zmw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/z"
	"github.com/gojekfarm/ziggurat/zb"
	"github.com/gojekfarm/ziggurat/zlog"
	"time"
)

var getCurrentTime = func() time.Time {
	return time.Now()
}

func MessageLogger(next z.MessageHandler) z.MessageHandler {
	return z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]interface{}{
			"ROUTE": messageEvent.StreamRoute,
			"VALUE": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		status := next.HandleMessage(messageEvent, app)
		switch status {
		case z.ProcessingSuccess:
			zlog.LogInfo("[Msg logger]: successfully processed message", args)
		case z.RetryMessage:
			zlog.LogInfo("[Msg logger]: retrying message", args)
		case z.SkipMessage:
			zlog.LogInfo("[Msg logger]: skipping message", args)
		}
		return status
	})
}

func MessageMetricsPublisher(next z.MessageHandler) z.MessageHandler {
	return z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]string{
			"topic_entity": messageEvent.StreamRoute,
			"kafka_topic":  messageEvent.Topic,
		}
		currTime := getCurrentTime()
		kafkaTimestamp := messageEvent.KafkaTimestamp
		delayInMS := currTime.Sub(kafkaTimestamp).Milliseconds()
		app.MetricPublisher().IncCounter("message_count", 1, args)
		app.MetricPublisher().Gauge("message_delay", delayInMS, args)
		return next.HandleMessage(messageEvent, app)
	})
}
