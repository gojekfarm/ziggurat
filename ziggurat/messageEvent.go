package ziggurat

import "github.com/rs/zerolog/log"

type MessageEvent struct {
	MessageKey         interface{}
	MessageValue       interface{}
	MessageValueBytes  []byte
	MessageKeyBytes    []byte
	ConsumerInstanceId string
	Topic              string
	Partition          int
	Attributes         map[string]interface{}
}

func NewMessageEvent(key []byte, value []byte, topic string, partition int, consumerInstanceId string) MessageEvent {
	return MessageEvent{
		MessageKeyBytes:    key,
		MessageValueBytes:  value,
		Topic:              topic,
		Partition:          partition,
		ConsumerInstanceId: consumerInstanceId,
	}
}

func (m *MessageEvent) GetMessageAttribute(key string) interface{} {
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	if m.Attributes[key] != nil {
		log.Warn().Msgf("key %s already exists and will be overwritten", key)
		m.Attributes[key] = value
	}
	m.Attributes[key] = value
}
