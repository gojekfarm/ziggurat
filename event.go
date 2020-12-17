package ziggurat

import (
	"sync"
	"time"
)

const StructVersion = "1.0"

type MessageEvent struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	StreamRoute       string
	ActualTimestamp   time.Time
	TimestampType     string
	Attributes        map[string]interface{}
	StructVersion     string
	//exposes Attributes for gob encoding, use GetAttribute and SetAttribute for thread safety
	mut *sync.Mutex
}

func (m *MessageEvent) PublishTime() time.Time {
	return m.ActualTimestamp
}

func (m *MessageEvent) MessageKey() []byte {
	return m.MessageKeyBytes
}

func (m *MessageEvent) MessageValue() []byte {
	return m.MessageValueBytes
}

func (m *MessageEvent) OriginTopic() string {
	return m.Topic
}

func (m *MessageEvent) RouteName() string {
	return m.StreamRoute
}

func (m *MessageEvent) GetAttribute(key string) interface{} {
	m.mut.Lock()
	defer m.mut.Unlock()
	return m.Attributes[key]
}

func (m *MessageEvent) SetAttribute(key string, value interface{}) {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.Attributes[key] = value
}

func (m *MessageEvent) Version() string {
	return m.StructVersion
}

func NewMessageEvent(key []byte, value []byte, topic string, route string, timestampType string, ktimestamp time.Time) *MessageEvent {
	return &MessageEvent{
		Attributes:        map[string]interface{}{},
		mut:               &sync.Mutex{},
		MessageValueBytes: value,
		MessageKeyBytes:   key,
		Topic:             topic,
		StreamRoute:       route,
		TimestampType:     timestampType,
		ActualTimestamp:   ktimestamp,
		StructVersion:     StructVersion,
	}
}
