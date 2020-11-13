package zig

import (
	"fmt"
	"sync"
	"time"
)

type Decoder func(model interface{}) error

type MessageEvent struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	KafkaTimestamp    time.Time
	DecodeValue       Decoder
	DecodeKey         Decoder
	TimestampType     string
	Attributes        map[string]interface{}
	amutex            *sync.Mutex
	// exposes Attributes for gob encoding, use Get and Set for thread safety
}

func NewMessageEvent(key []byte, value []byte, topic string, entity string, timestampType string, ktimestamp time.Time) MessageEvent {
	return MessageEvent{
		DecodeValue: func(model interface{}) error {
			return fmt.Errorf("no decoder found")
		},
		DecodeKey: func(model interface{}) error {
			return fmt.Errorf("no decoder found")
		},
		Attributes:        map[string]interface{}{},
		amutex:            &sync.Mutex{},
		MessageValueBytes: value,
		MessageKeyBytes:   key,
		Topic:             topic,
		TopicEntity:       entity,
		TimestampType:     timestampType,
		KafkaTimestamp:    ktimestamp,
	}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	m.amutex.Lock()
	defer m.amutex.Unlock()
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	m.amutex.Lock()
	defer m.amutex.Unlock()
	m.Attributes[key] = value
}
