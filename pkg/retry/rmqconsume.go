package retry

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/handler"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/streadway/amqp"
	"time"
)

func decodeMessage(body []byte) (basic.MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := basic.NewMessageEvent(nil, nil, "", "", "", time.Time{})
	if decodeErr := decoder.Decode(&messageEvent); decodeErr != nil {
		return messageEvent, decodeErr
	}
	return messageEvent, nil
}

func setupConsumers(app z.App, dialer *amqpextra.Dialer) error {
	handlerMap := app.Router().GetHandlerFunctionMap()
	serviceName := app.Config().ServiceName
	ctx := app.Context()

	for entity, topicEntity := range handlerMap {
		queueName := constructQueueName(serviceName, entity, QueueTypeInstant)
		messageHandler := handler.MessageHandler(app, topicEntity.HandlerFunc)
		consumerCTAG := fmt.Sprintf("%s_%s_%s", queueName, serviceName, "ctag")

		options := []consumer.Option{
			consumer.WithInitFunc(func(conn consumer.AMQPConnection) (consumer.AMQPChannel, error) {
				channel, err := conn.(*amqp.Connection).Channel()
				if err != nil {
					return nil, err
				}
				logger.LogError(channel.Qos(1, 0, false), "rmq consumer: error setting QOS", nil)
				return channel, nil
			}),
			consumer.WithContext(ctx),
			consumer.WithConsumeArgs(consumerCTAG, false, false, false, false, nil),
			consumer.WithQueue(queueName),
			consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp.Delivery) interface{} {
				logger.LogInfo("rmq consumer: processing message", map[string]interface{}{"queue-name": queueName})
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
			logger.LogError(fmt.Errorf("consumer closed"), "rmq consumer: closed", nil)

		}()
	}
	return nil
}
