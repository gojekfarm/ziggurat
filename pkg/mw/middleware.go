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
		logger.LogInfo("Msg logger middleware", args)
		return next(messageEvent, app)
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
