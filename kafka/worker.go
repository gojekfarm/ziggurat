package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
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
	consumer    *kafka.Consumer
	routeGroup  string
	pollTimeout int
	killSig     chan struct{}
	id          string
	err         error
}

func (w *worker) run(ctx context.Context) {
	defer func() {
		err := closeConsumer(w.consumer)
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
			ev := pollEvent(w.consumer, w.pollTimeout)
			switch e := ev.(type) {
			case *kafka.Message:
				processMessage(ctx, e, w.handler, w.logger, w.routeGroup)
			case kafka.Error:
				if e.IsFatal() {
					w.err = e
					run = false
				}
				// handle case `kakfa.Stats`
			default:
				// do nothing
			}
		}
	}
}

func (w *worker) kill() {
	w.killSig <- struct{}{}
}
