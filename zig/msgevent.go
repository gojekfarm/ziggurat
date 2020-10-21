package zig

import (
	"sync"
	"time"
)

type msgAttribute struct {
	*sync.RWMutex
	data map[string]interface{}
}

type MessageEvent struct {
	MessageKey        interface{}
	MessageValue      interface{}
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	KafkaTimestamp    time.Time
	TimestampType     string
	attributes        *msgAttribute
}

func newMessageAttribute() *msgAttribute {
	return &msgAttribute{
		RWMutex: &sync.RWMutex{},
		data:    map[string]interface{}{},
	}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	m.attributes.RLock()
	defer m.attributes.RUnlock()
	return m.attributes.data[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	m.attributes.RWMutex.Lock()
	defer m.attributes.RWMutex.Unlock()
	if m.attributes.data[key] != nil {
		m.attributes.data[key] = value
		return
	}
	m.attributes.data[key] = value
}
