package zig

import (
	"time"
)

type MessageEvent struct {
	MessageKey        interface{}
	MessageValue      interface{}
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	KafkaTimestamp    time.Time
	TimestampType     string
	Attributes        map[string]interface{}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	if m.Attributes[key] != nil {
		m.Attributes[key] = value
		return
	}
	m.Attributes[key] = value
}
