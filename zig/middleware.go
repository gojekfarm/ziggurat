package zig

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"time"
)

var getCurrTime = func() time.Time {
	return time.Now()
}

func JSONDecoder(next HandlerFunc) HandlerFunc {
	return func(message MessageEvent, app App) ProcessStatus {
		message.DecodeValue = func(model interface{}) error {
			return json.Unmarshal(message.MessageValueBytes, model)
		}
		return next(message, app)
	}
}

func MessageLogger(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app App) ProcessStatus {
		args := map[string]interface{}{
			"topic-entity":  messageEvent.TopicEntity,
			"kafka-topic":   messageEvent.Topic,
			"kafka-ts":      messageEvent.KafkaTimestamp.String(),
			"message-value": fmt.Sprintf("%s", messageEvent.MessageValueBytes),
		}
		logInfo("Msg logger middleware", args)
		return next(messageEvent, app)
	}
}

func ProtoDecoder(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app App) ProcessStatus {
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

func MessageMetricsPublisher(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app App) ProcessStatus {
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
