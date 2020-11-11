package zig

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

var getCurrTime = func() time.Time {
	return time.Now()
}

func JSONDeserializer(model interface{}) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(message MessageEvent, app App) ProcessStatus {
			messageValueBytes := message.MessageValueBytes
			typeModel := reflect.TypeOf(model)
			res := reflect.New(typeModel).Interface()
			if err := json.Unmarshal(messageValueBytes, res); err != nil {
				logError(err, "JSON middleware:", nil)
				message.MessageValue = nil
				return next(message, app)
			}
			message.MessageValue = res
			return next(message, app)
		}
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

func ProtobufDeserializer(protoMessage interface{}) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(messageEvent MessageEvent, app App) ProcessStatus {
			messageValueBytes := messageEvent.MessageValueBytes

			typeModel := reflect.TypeOf(protoMessage)
			res := reflect.New(typeModel).Interface()

			protoRes, ok := res.(proto.Message)

			if !ok {
				logError(ErrInterfaceNotProtoMessage, "Protobuf middleware", nil)
				return next(messageEvent, app)
			}
			if err := proto.Unmarshal(messageValueBytes, protoRes); err != nil {
				logError(err, "protobuf middleware unmarshal error", nil)
				return next(messageEvent, app)
			}
			messageEvent.MessageValue = protoRes
			return next(messageEvent, app)
		}
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
