package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

const defaultPollTimeoutInMS = 1000

var startConsumer = func(
				ctx context.Context,
				h ziggurat.Handler,
				l ziggurat.StructuredLogger,
				consumer *kafka.Consumer,
				route string, instanceID string,
				wg *sync.WaitGroup,
) {
	logChan := consumer.Logs()

	go func() {
		for evt := range logChan {
			if evt.Tag != "COMMIT" {
				l.Info(evt.Message, map[string]interface{}{
					"client":   evt.Name,
					"tag":      evt.Tag,
					"ts":       evt.Timestamp,
					"severity": evt.Level,
				})
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
				ev := pollEvent(consumer, defaultPollTimeoutInMS)
				switch e := ev.(type) {
				case *kafka.Message:
					// blocks until process returns
					processMessage(ctx, e, consumer, h, l, route)
				case kafka.Error:
					l.Error("kafka poll error", e)
				default:
					//Do nothing
				}
			}
		}
		l.Error("stopping consumer", ctx.Err(), map[string]interface{}{"consumerID": instanceID})
		wg.Done()
	}(instanceID)
}

var StartConsumers = func(
				ctx context.Context,
				consumerConfig *kafka.ConfigMap,
				route string,
				topics []string,
				instances int,
				h ziggurat.Handler,
				l ziggurat.StructuredLogger,
				wg *sync.WaitGroup,
) []*kafka.Consumer {
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
