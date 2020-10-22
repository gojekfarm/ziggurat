package zig

import (
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

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
	// exposes Attributes for gob encoding, use Get and Set for thread safety
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	m.Attributes[key] = value
}
