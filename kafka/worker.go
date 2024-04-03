package kafka

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ErrorWorkerKilled struct {
	workerID string
}

func (e ErrorWorkerKilled) Error() string {
	return fmt.Sprintf("%s consumer killed", e.workerID)
}

type worker struct {
	handler     ziggurat.Handler
	logger      ziggurat.StructuredLogger
	consumer    confluentConsumer
	routeGroup  string
	pollTimeout int
	killSig     chan struct{}
	id          string
	err         error
}

func (w *worker) run(ctx context.Context) {

	defer func() {
		_, err := w.consumer.Commit()
		if err != nil {
			w.logger.Error("pre-close commit error", err, map[string]interface{}{"Worker-ID": w.id})
		}
		err = w.consumer.Close()
		w.logger.Error("error closing kafka consumer", err, map[string]interface{}{"Worker-ID": w.id})
	}()

	lch := w.consumer.Logs()
	go func() {
		for evt := range lch {
			w.logger.Info(evt.Message, map[string]interface{}{
				"client": evt.Name,
				"lvl":    evt.Level,
			})
		}
	}()

	done := ctx.Done()
	run := true

	for run {
		select {
		case <-done:
			w.err = ctx.Err()
			run = false
		case <-w.killSig:
			w.err = ErrorWorkerKilled{workerID: w.id}
			run = false
		default:
			ev := w.consumer.Poll(w.pollTimeout)
			switch e := ev.(type) {
			case *kafka.Message:
				processMessage(ctx, e, w.handler, w.routeGroup)
				if err := storeOffsets(w.consumer, e.TopicPartition); err != nil {
					w.logger.Error("error storing offsets locally", err)
				}
			case kafka.Error:
				if e.IsFatal() {
					w.err = e
					run = false
				}
				w.logger.Error("kafka poll error", e)
				// handle case `kafka.Stats`
			default:
				// do nothing
			}
		}
	}
}

func (w *worker) kill() {
	w.killSig <- struct{}{}
}
