package zmw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"time"
)

var getCurrentTime = func() time.Time {
	return time.Now()
}

func MessageLogger(next ztype.MessageHandler) ztype.MessageHandler {
	return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		args := map[string]interface{}{
			"ROUTE": messageEvent.StreamRoute,
			"VALUE": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		status := next.HandleMessage(messageEvent, app)
		switch status {
		case ztype.ProcessingSuccess:
			zlog.LogInfo("[Msg logger]: successfully processed message", args)
		case ztype.RetryMessage:
			zlog.LogInfo("[Msg logger]: retrying message", args)
		case ztype.SkipMessage:
			zlog.LogInfo("[Msg logger]: skipping message", args)
		}
		return status
	})
}

func MessageMetricsPublisher(next ztype.MessageHandler) ztype.MessageHandler {
	return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
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
