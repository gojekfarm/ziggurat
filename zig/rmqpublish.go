package zig

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/streadway/amqp"
	amqpsafe "github.com/xssnick/amqp-safe"
	"time"
)

const RetryCount = "retryCount"

func getRetryCount(m *MessageEvent) int {
	if value := m.GetMessageAttribute(RetryCount); value == nil {
		return 0
	}
	return m.GetMessageAttribute(RetryCount).(int)
}

func setRetryCount(m *MessageEvent) {
	value := m.GetMessageAttribute(RetryCount)

	if value == nil {
		m.SetMessageAttribute(RetryCount, 1)
		return
	}
	m.SetMessageAttribute(RetryCount, value.(int)+1)
}

func publishMessage(ctx context.Context, c *amqpsafe.Connector, exchangeName string, payload MessageEvent, expirationInMS string) error {
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

	doneCh := ctx.Done()
	for {
		select {
		case <-doneCh:
			return nil
		default:
			if publishErr := c.Publish(exchangeName, "", publishing); publishErr != nil {
				if publishErr == amqpsafe.ErrNoChannel {
					logError(publishErr, "rmq publish: retrying publish", map[string]interface{}{"exchange-name": exchangeName})
					time.Sleep(2 * time.Second)
					continue
				}
				return publishErr
			}
		}
		return nil
	}
}

func retry(ctx context.Context, c *amqpsafe.Connector, config *Config, payload MessageEvent, expiry string) error {
	exchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDelay)
	deadLetterExchangeName := constructExchangeName(config.ServiceName, payload.TopicEntity, QueueTypeDL)
	retryCount := getRetryCount(&payload)
	if retryCount == config.Retry.Count {
		return publishMessage(ctx, c, deadLetterExchangeName, payload, "")
	}
	setRetryCount(&payload)
	return publishMessage(ctx, c, exchangeName, payload, expiry)
}
