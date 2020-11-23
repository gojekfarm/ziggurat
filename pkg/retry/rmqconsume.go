package retry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/handler"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/makasim/amqpextra"
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

var setupConsumers = func(app z.App, dialer *amqpextra.Dialer) error {
	handlerMap := app.Router().HandlerFuncEntityMap()
	serviceName := app.Config().ServiceName

	for entity, handlerFunc := range handlerMap {
		queueName := constructQueueName(serviceName, entity, QueueTypeInstant)
		messageHandler := handler.MessageHandlerMW(handlerFunc)
		consumerCTAG := fmt.Sprintf("%s_%s_%s", queueName, serviceName, "ctag")

		c, err := createConsumer(app, dialer, consumerCTAG, queueName, messageHandler)

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
