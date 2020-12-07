package retry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"github.com/gojekfarm/ziggurat-go/pkg/zlog"
	"github.com/gojekfarm/ziggurat-go/pkg/zmw"
	"github.com/makasim/amqpextra"
	"time"
)

var decodeMessage = func(body []byte) (zb.MessageEvent, error) {
	buff := bytes.Buffer{}
	buff.Write(body)
	decoder := gob.NewDecoder(&buff)
	messageEvent := zb.NewMessageEvent(nil, nil, "", "", "", time.Time{})
	if decodeErr := decoder.Decode(&messageEvent); decodeErr != nil {
		return messageEvent, decodeErr
	}
	return messageEvent, nil
}

var setupConsumers = func(app z.App, dialer *amqpextra.Dialer) error {
	routes := app.Routes()
	messageHandler := app.Handler()
	serviceName := app.ConfigStore().Config().ServiceName

	for _, route := range routes {
		queueName := constructQueueName(serviceName, route, QueueTypeInstant)
		messageHandler := zmw.DefaultTerminalMW(messageHandler)
		consumerCTAG := fmt.Sprintf("%s_%s_%s", queueName, serviceName, "ctag")

		c, err := createConsumer(app, dialer, consumerCTAG, queueName, messageHandler)

		if err != nil {
			return err
		}
		go func() {
			<-c.NotifyClosed()
			zlog.LogError(fmt.Errorf("consumer closed"), "rmq consumer: closed", nil)
		}()
	}
	return nil
}