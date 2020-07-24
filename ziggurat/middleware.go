package ziggurat

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

func JSONDeserializer(handlerFn HandlerFunc, jsonValue interface{}) HandlerFunc {
	return func(message interface{}) {
		if m, ok := message.(*kafka.Message); ok {
			err := json.Unmarshal(m.Value, jsonValue)
			handlerFn(jsonValue)
			if err != nil {
				log.Printf("error unmarshalling JSON %v", err)
			}
		} else {
			log.Printf("middleware execution failed: expected *kafka.Message but got %T", message)
		}

	}
}
