package mw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"time"
)

var getCurrentTime = func() time.Time {
	return time.Now()
}

func MessageLogger(next z.MessageHandler) z.MessageHandler {
	return z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]interface{}{
			"topic-entity":  messageEvent.StreamRoute,
			"kafka-topic":   messageEvent.Topic,
			"kafka-ts":      messageEvent.KafkaTimestamp.String(),
			"message-value": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		status := next.HandleMessage(messageEvent, app)
		switch status {
		case z.ProcessingSuccess:
			logger.LogInfo("Msg logger middleware: successfully processed message", args)
		case z.RetryMessage:
			logger.LogInfo("Msg logger middleware: retrying message", args)
		case z.SkipMessage:
			logger.LogInfo("Msg logger middleware: skipping message", args)
		}
		return status
	})
}

func MessageMetricsPublisher(next z.MessageHandler) z.MessageHandler {
	return z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
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
