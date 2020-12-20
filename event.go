package ziggurat

import (
	"sync"
)

type MsgAttributes map[string]interface{}

type Message struct {
	Value      []byte
	Key        []byte
	Attributes MsgAttributes
	RouteName  string
	attrMutex  *sync.Mutex
	//exposes Attributes for gob encoding, use Get and Set for thread safety
}

func (m Message) Attribute(key string) interface{} {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	return m.Attributes[key]
}

func (m *Message) SetAttribute(key string, value interface{}) {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	m.Attributes[key] = value
}

func CreateMessage(key []byte, value []byte, routeName string, attributes map[string]interface{}) Message {
	m := Message{
		Value:      value,
		Key:        key,
		Attributes: attributes,
		RouteName:  routeName,
		attrMutex:  &sync.Mutex{},
	}
	if m.Attributes == nil {
		m.Attributes = map[string]interface{}{}
	}
	return m
}
