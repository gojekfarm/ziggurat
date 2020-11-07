package zig

import (
	"bytes"
	"encoding/gob"
	"github.com/rs/zerolog/log"
	amqpsafe "github.com/xssnick/amqp-safe"
)

func decodeMessage(body []byte) (MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := &MessageEvent{Attributes: map[string]interface{}{}}
	if decodeErr := decoder.Decode(messageEvent); decodeErr != nil {
		return *messageEvent, decodeErr
	}
	return *messageEvent, nil
}

func createSetupCallback(connector *amqpsafe.Connector, app App) func() {
	topicEntities := app.Router().GetTopicEntityNames()
	handlerMap := app.Router().GetHandlerFunctionMap()
	return func() {
		declareExchanges(connector, topicEntities, app.Config().ServiceName)
		createInstantQueues(connector, topicEntities, app.Config().ServiceName)
		createDelayQueues(connector, topicEntities, app.Config().ServiceName)
		createDeadLetterQueues(connector, topicEntities, app.Config().ServiceName)
		for _, te := range topicEntities {
			queueName := constructQueueName(app.Config().ServiceName, te, QueueTypeInstant)
			connector.Consume(queueName, queueName+"_consumer", func(body []byte) amqpsafe.Result {
				msg, err := decodeMessage(body)
				if err != nil {
					return amqpsafe.ResultReject
				}
				MessageHandler(app, handlerMap[te].HandlerFunc)(msg)
				log.Info().Msg("[RABBITMQ CONSUMER] message consumed successfully")
				return amqpsafe.ResultOK
			})
		}
	}
}
