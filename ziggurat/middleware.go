package ziggurat

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

func JSONDeserializer(handlerFn HandlerFunc, structValue interface{}) HandlerFunc {
	return func(message MessageEvent) ProcessStatus {
		messageValueBytes := message.MessageValueBytes
		if err := json.Unmarshal(messageValueBytes, structValue); err != nil {
			log.Error().Err(err)
		}
		message.MessageValue = structValue
		return handlerFn(message)
	}
}

func ProtobufDeserializer(handler HandlerFunc, messageValue proto.Message) HandlerFunc {
	return func(messageEvent MessageEvent) ProcessStatus {
		messageValueBytes := messageEvent.MessageValueBytes
		if err := proto.Unmarshal(messageValueBytes, messageValue); err != nil {
			messageEvent.MessageValue = messageValue
			log.Error().Err(err)
		}
		messageEvent.MessageValue = messageValue
		return handler(messageEvent)
	}
}
