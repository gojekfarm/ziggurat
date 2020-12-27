package ziggurat

import (
	"context"
	"time"
)

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event struct {
	Value     []byte
	ctx       context.Context
	Headers   map[string]string
	Timestamp time.Time
}

func (m Event) Context() context.Context {
	return m.ctx
}

func CreateMessageEvent(value []byte, headers map[string]string, timestamp time.Time, ctx context.Context) Event {
	return Event{
		Value:     value,
		ctx:       ctx,
		Headers:   headers,
		Timestamp: timestamp,
	}
}
