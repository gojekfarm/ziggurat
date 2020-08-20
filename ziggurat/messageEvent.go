package ziggurat

import "github.com/rs/zerolog/log"

type MessageEvent struct {
	MessageKey        interface{}
	MessageValue      interface{}
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	attributes        map[string]interface{}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	return m.attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	if m.attributes[key] != nil {
		log.Warn().Msgf("key %s already exists and will be overwritten", key)
		m.attributes[key] = value
	}
	m.attributes[key] = value
}
