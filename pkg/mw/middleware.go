package mw

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"time"
)

var GetCurrTime = func() time.Time {
	return time.Now()
}

func MessageLogger(next z.HandlerFunc) z.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]interface{}{
			"topic-entity":  messageEvent.TopicEntity,
			"kafka-topic":   messageEvent.Topic,
			"kafka-ts":      messageEvent.KafkaTimestamp.String(),
			"message-value": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		status := next(messageEvent, app)
		switch status {
		case z.ProcessingSuccess:
			logger.LogInfo("Msg logger middleware: successfully processed message", args)
		case z.RetryMessage:
			logger.LogInfo("Msg logger middleware: retrying message", args)
		}
		return status
	}
}

func MessageMetricsPublisher(next z.HandlerFunc) z.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]string{
			"topic_entity": messageEvent.TopicEntity,
			"kafka_topic":  messageEvent.Topic,
		}
		currTime := GetCurrTime()
		kafkaTimestamp := messageEvent.KafkaTimestamp
		delayInMS := currTime.Sub(kafkaTimestamp).Milliseconds()
		app.MetricPublisher().IncCounter("message_count", 1, args)
		app.MetricPublisher().Gauge("message_delay", delayInMS, args)
		return next(messageEvent, app)
	}
}
