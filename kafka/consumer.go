package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"sync"
	"time"
)

const defaultPollTimeout = 1 * time.Second
const brokerRetryTimeout = 3 * time.Second

var startConsumer = func(ctx context.Context, h ziggurat.Handler, l ziggurat.StructuredLogger, consumer *kafka.Consumer, route string, instanceID string, wg *sync.WaitGroup) {
	logChan := consumer.Logs()

	go func() {
		for evt := range logChan {
			if evt.Tag != "COMMIT" {
				l.Info(evt.Message, map[string]interface{}{"client": evt.Name, "tag": evt.Tag, "ts": evt.Timestamp, "severity": evt.Level})
			}
		}
	}()

	go func(instanceID string) {
		run := true
		doneCh := ctx.Done()

		for run {
			select {
			case <-doneCh:
				run = false
			default:
				msg, err := readMessage(consumer, defaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					time.Sleep(brokerRetryTimeout)
					continue
				}
				if msg != nil {
					message := Message{
						headers: map[string]string{ziggurat.HeaderMessageRoute: route, ziggurat.HeaderMessageType: "kafka"},
						value:   msg.Value,
					}
					h.HandleEvent(message, ctx)
				}
			}
		}
		l.Error("stopping consumer", ctx.Err(), map[string]interface{}{"consumerID": instanceID})
		wg.Done()
	}(instanceID)
}

var StartConsumers = func(ctx context.Context, consumerConfig *kafka.ConfigMap, route string, topics []string, instances int, h ziggurat.Handler, l ziggurat.StructuredLogger, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, l, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s_%s_%d", route, groupID, i)
		wg.Add(1)
		startConsumer(ctx, h, l, consumer, route, instanceID, wg)
	}
	return consumers
}
