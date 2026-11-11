package kafka

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"

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

// exitReason maps the error a worker stopped with to a WorkerExitsCounter reason.
func exitReason(err error) string {
	var killed ErrorWorkerKilled
	switch {
	case errors.As(err, &killed):
		return prometheus.WorkerExitKilled
	case err == nil, errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return prometheus.WorkerExitContextDone
	default:
		return prometheus.WorkerExitFatalError
	}
}

func (w *worker) run(ctx context.Context) {
	prometheus.WorkersRunningGauge.WithLabelValues(w.routeGroup).Inc()
	lag := newLagReporter(w.routeGroup)
	defer func() {
		lag.clear()
		prometheus.WorkersRunningGauge.WithLabelValues(w.routeGroup).Dec()
		prometheus.WorkerExitsCounter.WithLabelValues(w.routeGroup, exitReason(w.err)).Inc()
	}()
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
			w.logger.Debug(evt.Message, map[string]any{
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
			case kafka.Stats:
				if err := lag.update(e.String()); err != nil {
					w.logger.Error("kafka statistics error", err, map[string]interface{}{"Worker-ID": w.id})
				}
			case kafka.Error:
				prometheus.KafkaErrorsCounter.WithLabelValues(w.routeGroup, e.Code().String(), strconv.FormatBool(e.IsFatal())).Inc()
				if e.IsFatal() {
					w.err = e
					run = false
				}
				w.logger.Error("kafka poll error", e)
			default:
				// do nothing
			}
		}
	}
}

func (w *worker) kill() {
	w.killSig <- struct{}{}
}
