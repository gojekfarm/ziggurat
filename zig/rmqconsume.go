package zig

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/streadway/amqp"
	"sync"
)

func decodeMessage(body []byte) (MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := &MessageEvent{Attributes: map[string]interface{}{},
		DecodeValue: func(model interface{}) error { return nil },
		DecodeKey:   func(model interface{}) error { return nil },
		amutex:      &sync.Mutex{},
	}
	if decodeErr := decoder.Decode(messageEvent); decodeErr != nil {
		return *messageEvent, decodeErr
	}
	return *messageEvent, nil
}

func setupConsumers(app App, dialer *amqpextra.Dialer) error {
	handlerMap := app.Router().GetHandlerFunctionMap()
	serviceName := app.Config().ServiceName
	ctx := app.Context()

	for entity, topicEntity := range handlerMap {
		queueName := constructQueueName(serviceName, entity, QueueTypeInstant)
		messageHandler := MessageHandler(app, topicEntity.HandlerFunc)
		consumerCTAG := fmt.Sprintf("%s_%s_%s", queueName, serviceName, "ctag")

		options := []consumer.Option{
			consumer.WithInitFunc(func(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
				channel, err := conn.(*amqp.Connection).Channel()
				if err != nil {
					return nil, err
				}
				logError(channel.Qos(1, 0, false), "rmq consumer: error setting QOS", nil)
				return channel, nil
			}),
			consumer.WithContext(ctx),
			consumer.WithConsumeArgs(consumerCTAG, false, false, false, false, nil),
			consumer.WithQueue(queueName),
			consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
				logInfo("rmq consumer: processing message", map[string]interface{}{"queue-name": queueName})
				msgEvent, err := decodeMessage(msg.Body)
				if err != nil {
					return msg.Reject(true)
				}
				messageHandler(msgEvent)
				return msg.Ack(false)
			}))}
		c, err := dialer.Consumer(options...)

		if err != nil {
			return err
		}
		go func() {
			<-c.NotifyClosed()
			logError(fmt.Errorf("consumer closed"), "rmq consumer: closed", nil)

		}()
	}
	return nil
}
