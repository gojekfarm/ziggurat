package zretry

import (
	"bytes"
	"encoding/gob"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

const retryCount = "retryCount"

func GetRetryCount(m *zbase.MessageEvent) int {
	if value := m.GetMessageAttribute(retryCount); value == nil {
		return 0
	}
	return m.GetMessageAttribute(retryCount).(int)
}

func SetRetryCount(m *zbase.MessageEvent) {
	value := m.GetMessageAttribute(retryCount)

	if value == nil {
		m.SetMessageAttribute(retryCount, 1)
		return
	}
	m.SetMessageAttribute(retryCount, value.(int)+1)
}

var PublishMessage = func(exchangeName string, p *publisher.Publisher, payload zbase.MessageEvent, expirationInMS string) error {
	buff := bytes.Buffer{}
	encoder := gob.NewEncoder(&buff)
	if encodeErr := encoder.Encode(payload); encodeErr != nil {
		return encodeErr
	}
	publishing := amqp.Publishing{
		Body:        buff.Bytes(),
		ContentType: "text/plain",
	}
	if expirationInMS != "" {
		publishing.Expiration = expirationInMS
	}
	err := p.Publish(publisher.Message{
		Exchange:   exchangeName,
		Publishing: publishing,
	})
	return err
}

func retry(p *publisher.Publisher, prefix string, retryCount int, payload zbase.MessageEvent, expiry string) error {
	exchangeName := constructExchangeName(prefix, payload.StreamRoute, QueueTypeDelay)
	deadLetterExchangeName := constructExchangeName(prefix, payload.StreamRoute, QueueTypeDL)
	retryCountPayload := GetRetryCount(&payload)
	if retryCountPayload == retryCount {
		return PublishMessage(deadLetterExchangeName, p, payload, "")
	}
	SetRetryCount(&payload)
	return PublishMessage(exchangeName, p, payload, expiry)
}
