package zig

import (
	"bytes"
	"encoding/gob"
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

func createSetupCallback(consConn *amqpsafe.Connector, app App) func() {
	topicEntities := app.Router().GetTopicEntityNames()
	handlerMap := app.Router().GetHandlerFunctionMap()
	return func() {
		declareExchanges(consConn, topicEntities, app.Config().ServiceName)
		createInstantQueues(consConn, topicEntities, app.Config().ServiceName)
		createDelayQueues(consConn, topicEntities, app.Config().ServiceName)
		createDeadLetterQueues(consConn, topicEntities, app.Config().ServiceName)
		for _, te := range topicEntities {
			queueName := constructQueueName(app.Config().ServiceName, te, QueueTypeInstant)
			consConn.Consume(queueName, queueName+"_consumer", func(body []byte) amqpsafe.Result {
				msg, err := decodeMessage(body)
				if err != nil {
					logError(err, "ziggurat rmq consumer: message decode error", map[string]interface{}{"topic-entity": te})
					return amqpsafe.ResultError
				}
				MessageHandler(app, handlerMap[te].HandlerFunc)(msg)
				logInfo("ziggurat rmq consumer: processed message successfully", map[string]interface{}{"topic-entity": te})
				return amqpsafe.ResultOK
			})
		}
	}
}
