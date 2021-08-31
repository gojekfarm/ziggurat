package kafka

import (
	"context"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

const logTagToSkip = "COMMIT"
const defaultPollTimeoutInMS = 1000

var startConsumer = func(
	ctx context.Context,
	h ziggurat.Handler,
	l ziggurat.StructuredLogger,
	consumer *kafka.Consumer,
	route string,
	cfgMap kafka.ConfigMap,
	wg *sync.WaitGroup,
) {
	logChan := consumer.Logs()

	go func() {
		for evt := range logChan {
			if evt.Tag != logTagToSkip {
				l.Info(evt.Message, map[string]interface{}{
					"client": evt.Name,
					"lvl":  evt.Level,
				})
			}
		}
	}()

	go func() {
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
					processMessage(ctx, e, consumer, h, l, cfgMap, route)
				case kafka.Error:
					l.Error("kafka poll error", e)
				default:
					//Do nothing
					// Expose metrics in future releases
				}
			}
		}
		l.Error("stopping consumer", ctx.Err(), map[string]interface{}{"client": consumer.String()})
		wg.Done()
	}()
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
		wg.Add(1)
		startConsumer(ctx, h, l, consumer, route, *consumerConfig, wg)
	}
	return consumers
}
