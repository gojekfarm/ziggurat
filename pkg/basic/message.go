package basic

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"sync"
	"time"
)

type DecoderHook func(model interface{}) error

type MessageEvent struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	Topic             string
	TopicEntity       string
	KafkaTimestamp    time.Time
	ValueDecoderHook  DecoderHook
	KeyDecoderHook    DecoderHook
	TimestampType     string
	Attributes        map[string]interface{}
	attrMutex         *sync.Mutex
	//exposes Attributes for gob encoding, use Get and Set for thread safety
}

func NewMessageEvent(key []byte, value []byte, topic string, entity string, timestampType string, ktimestamp time.Time) MessageEvent {
	return MessageEvent{
		ValueDecoderHook: func(model interface{}) error {
			return zerror.ErrNoDecoderFound
		},
		KeyDecoderHook: func(model interface{}) error {
			return zerror.ErrNoDecoderFound
		},
		Attributes:        map[string]interface{}{},
		attrMutex:         &sync.Mutex{},
		MessageValueBytes: value,
		MessageKeyBytes:   key,
		Topic:             topic,
		TopicEntity:       entity,
		TimestampType:     timestampType,
		KafkaTimestamp:    ktimestamp,
	}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	m.Attributes[key] = value
}
