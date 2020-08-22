package ziggurat

import (
	"github.com/rs/zerolog/log"
)

type MessageEvent struct {
	MessageKey        interface{}
	MessageValue      interface{}
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	Attributes        map[string]interface{}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	if m.Attributes[key] != nil {
		log.Warn().Msgf("key %s already exists and will be overwritten", key)
		m.Attributes[key] = value
		return
	}
	m.Attributes[key] = value
}
