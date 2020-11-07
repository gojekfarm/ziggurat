package ziggurat

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"reflect"
	"time"
)

func JSONDeserializer(model interface{}) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(message MessageEvent, app App) ProcessStatus {
			messageValueBytes := message.MessageValueBytes
			typeModel := reflect.TypeOf(model)
			res := reflect.New(typeModel).Interface()
			if err := json.Unmarshal(messageValueBytes, res); err != nil {
				log.Error().Err(err).Msg("[JSON MIDDLEWARE]")
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
		log.Info().
			Str("topic-entity", messageEvent.TopicEntity).
			Str("kafka-topic", messageEvent.Topic).
			Str("kafka-time-stamp", messageEvent.KafkaTimestamp.String()).
			Str("message-value", string(messageEvent.MessageValueBytes)).
			Msg("[MESSAGE LOGGER MIDDLEWARE]")
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
				log.Error().Err(ErrInterfaceNotProtoMessage).Msg("[PROTOBUF-MIDDLEWARE]")
				return next(messageEvent, app)
			}
			if err := proto.Unmarshal(messageValueBytes, protoRes); err != nil {
				log.Error().Err(err).Msg("[PROTOBUF-MIDDLEWARE]")
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
		currTime := time.Now()
		kafkaTimestamp := messageEvent.KafkaTimestamp
		delayInMS := currTime.Sub(kafkaTimestamp).Milliseconds()
		app.MetricPublisher().IncCounter("message_count", 1, args)
		app.MetricPublisher().Gauge("message_delay", delayInMS, args)
		return next(messageEvent, app)
	}
}
