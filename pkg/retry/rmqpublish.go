package retry

import (
	"bytes"
	"encoding/gob"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

const retryCount = "retryCount"

func getRetryCount(m *basic.MessageEvent) int {
	if value := m.GetMessageAttribute(retryCount); value == nil {
		return 0
	}
	return m.GetMessageAttribute(retryCount).(int)
}

func setRetryCount(m *basic.MessageEvent) {
	value := m.GetMessageAttribute(retryCount)

	if value == nil {
		m.SetMessageAttribute(retryCount, 1)
		return
	}
	m.SetMessageAttribute(retryCount, value.(int)+1)
}

func publishMessage(exchangeName string, p *publisher.Publisher, payload basic.MessageEvent, expirationInMS string) error {
	defer p.Close()
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

func retry(p *publisher.Publisher, config *basic.Config, payload basic.MessageEvent, expiry string) error {
	exchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDelay)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDL)
	retryCount := getRetryCount(&payload)
	if retryCount == config.Retry.Count {
		return publishMessage(deadLetterExchangeName, p, payload, "")
	}
	setRetryCount(&payload)
	return publishMessage(exchangeName, p, payload, expiry)
}
