package retry

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

const RetryCount = "retryCount"

func getRetryCount(m *basic.MessageEvent) int {
	if value := m.GetMessageAttribute(RetryCount); value == nil {
		return 0
	}
	return m.GetMessageAttribute(RetryCount).(int)
}

func setRetryCount(m *basic.MessageEvent) {
	value := m.GetMessageAttribute(RetryCount)

	if value == nil {
		m.SetMessageAttribute(RetryCount, 1)
		return
	}
	m.SetMessageAttribute(RetryCount, value.(int)+1)
}

func publishMessage(ctx context.Context, exchangeName string, p *publisher.Publisher, payload basic.MessageEvent, expirationInMS string) error {
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

func retry(ctx context.Context, p *publisher.Publisher, config *basic.Config, payload basic.MessageEvent, expiry string) error {
	exchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDelay)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDL)
	retryCount := getRetryCount(&payload)
	if retryCount == config.Retry.Count {
		return publishMessage(ctx, deadLetterExchangeName, p, payload, "")
	}
	setRetryCount(&payload)
	return publishMessage(ctx, exchangeName, p, payload, expiry)
}
