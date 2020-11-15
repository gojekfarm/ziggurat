package mw

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/z"
	"github.com/golang/protobuf/proto"
	"time"
)

var getCurrTime = func() time.Time {
	return time.Now()
}

func JSONDecoder(next z.HandlerFunc) z.HandlerFunc {
	return func(message basic.MessageEvent, app z.App) z.ProcessStatus {
		message.DecodeValue = func(model interface{}) error {
			return json.Unmarshal(message.MessageValueBytes, model)
		}
		return next(message, app)
	}
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

func ProtoDecoder(next z.HandlerFunc) z.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		messageEvent.DecodeValue = func(model interface{}) error {
			if protoMessage, ok := model.(proto.Message); !ok {
				return errors.New("model not of type `proto.Message`")
			} else {
				return proto.Unmarshal(messageEvent.MessageValueBytes, protoMessage)
			}
		}
		return next(messageEvent, app)
	}
}

func MessageMetricsPublisher(next z.HandlerFunc) z.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		args := map[string]string{
			"topic_entity": messageEvent.TopicEntity,
			"kafka_topic":  messageEvent.Topic,
		}
		currTime := getCurrTime()
		kafkaTimestamp := messageEvent.KafkaTimestamp
		delayInMS := currTime.Sub(kafkaTimestamp).Milliseconds()
		app.MetricPublisher().IncCounter("message_count", 1, args)
		app.MetricPublisher().Gauge("message_delay", delayInMS, args)
		return next(messageEvent, app)
	}
}
