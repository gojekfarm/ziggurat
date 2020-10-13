package zig

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
)

func JSONDeserializer(structValue interface{}) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(message MessageEvent, app *App) ProcessStatus {
			messageValueBytes := message.MessageValueBytes
			if err := json.Unmarshal(messageValueBytes, structValue); err != nil {
				log.Error().Err(err).Msg("error de-serialising json")
				message.MessageValue = nil
				return next(message, app)
			}
			message.MessageValue = structValue
			return next(message, app)
		}
	}
}

func MessageLogger(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *App) ProcessStatus {
		log.Info().
			Str("topic-entity", messageEvent.TopicEntity).
			Str("kafka-topic", messageEvent.Topic).
			Str("kafka-time-stamp", messageEvent.KafkaTimestamp.String()).
			Str("message-value", string(messageEvent.MessageValueBytes)).
			Msg("received message")
		return next(messageEvent, app)
	}
}

func ProtobufDeserializer(messageValue proto.Message) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(messageEvent MessageEvent, app *App) ProcessStatus {
			messageValueBytes := messageEvent.MessageValueBytes
			if err := proto.Unmarshal(messageValueBytes, messageValue); err != nil {
				messageEvent.MessageValue = messageValue
				log.Error().Err(err).Msg("error de-serializing protobuf")
				messageEvent.MessageValue = nil
				return next(messageEvent, app)
			}
			messageEvent.MessageValue = messageValue
			return next(messageEvent, app)
		}
	}
}
