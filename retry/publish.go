package retry

import (
	"bytes"
	"encoding/gob"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

const retryCount = "retryCount"

func getRetryCount(m *zbase.MessageEvent) int {
	if value := m.GetMessageAttribute(retryCount); value == nil {
		return 0
	}
	return m.GetMessageAttribute(retryCount).(int)
}

func setRetryCount(m *zbase.MessageEvent) {
	value := m.GetMessageAttribute(retryCount)

	if value == nil {
		m.SetMessageAttribute(retryCount, 1)
		return
	}
	m.SetMessageAttribute(retryCount, value.(int)+1)
}

var publishMessage = func(exchangeName string, p *publisher.Publisher, payload zbase.MessageEvent, expirationInMS string) error {
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

func retry(p *publisher.Publisher, config *zbase.Config, payload zbase.MessageEvent, expiry string) error {
	exchangeName := constructExchangeName(config.ServiceName, payload.StreamRoute, QueueTypeDelay)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.StreamRoute, QueueTypeDL)
	retryCount := getRetryCount(&payload)
	if retryCount == config.Retry.Count {
		return publishMessage(deadLetterExchangeName, p, payload, "")
	}
	setRetryCount(&payload)
	return publishMessage(exchangeName, p, payload, expiry)
}
