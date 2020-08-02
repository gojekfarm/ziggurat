package ziggurat
//
//import (
//	"encoding/json"
//	"github.com/confluentinc/confluent-kafka-go/kafka"
//	"github.com/rs/zerolog/log"
//)
//
//func JSONDeserializer(handlerFn HandlerFunc, structValue interface{}) HandlerFunc {
//	return func(message interface{}) {
//		if m, ok := message.(*kafka.Message); ok {
//			err := json.Unmarshal(m.Value, structValue)
//			handlerFn(structValue)
//			if err != nil {
//				log.Error().Err(err)
//			}
//		} else {
//			log.Error().Msgf("middleware execution failed: expected *kafka.Message but got %T", message)
//		}
//	}
//}
