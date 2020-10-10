package ziggurat

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
)

func JSONDeserializer(handlerFn HandlerFunc, structValue interface{}) HandlerFunc {
	return func(message MessageEvent, app *App) ProcessStatus {
		messageValueBytes := message.MessageValueBytes
		if err := json.Unmarshal(messageValueBytes, structValue); err != nil {
			log.Error().Err(err).Msg("error deserializing json")
			message.MessageValue = nil
			return handlerFn(message, app)
		}
		message.MessageValue = structValue
		return handlerFn(message, app)
	}
}

func ProtobufDeserializer(handler HandlerFunc, messageValue proto.Message) HandlerFunc {
	return func(messageEvent MessageEvent, app *App) ProcessStatus {
		messageValueBytes := messageEvent.MessageValueBytes
		if err := proto.Unmarshal(messageValueBytes, messageValue); err != nil {
			messageEvent.MessageValue = messageValue
			log.Error().Err(err).Msg("error deserializing protobuf")
			messageEvent.MessageValue = nil
			return handler(messageEvent, app)
		}
		messageEvent.MessageValue = messageValue
		return handler(messageEvent, app)
	}
}
