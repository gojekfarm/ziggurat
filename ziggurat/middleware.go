package ziggurat

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
)

func JSONDeserializer(handlerFn HandlerFunc, structValue interface{}) HandlerFunc {
	return func(message MessageEvent) {
		messageValueBytes := message.MessageValueBytes
		if err := json.Unmarshal(messageValueBytes, structValue); err != nil {
			log.Error().Err(err)
		}
		message.MessageValue = structValue
		handlerFn(message)
	}
}
