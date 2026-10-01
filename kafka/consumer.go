package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"sync"
)

var ErrCleanShutdown = errors.New("error: clean shutdown of kafka consumers")

type ConsumerGroup struct {
	workers          []*worker
	Logger           ziggurat.StructuredLogger
	GroupConfig      ConsumerConfig
	wg               *sync.WaitGroup
	c                confluentConsumer
	consumerMakeFunc func(*kafka.ConfigMap, []string) confluentConsumer
}

func (cg *ConsumerGroup) Consume(ctx context.Context, handler ziggurat.Handler) error {

	cg.init()

	if cg.GroupConfig.ConsumerCount < 1 {
		cg.Logger.Warn("ConsumerConfig.GroupCount < 1, no consumers will be started")
	}

	grpConfig := cg.GroupConfig
	groupID := grpConfig.GroupID
	// sets default pollTimeout of 100ms
	pollTimeout := 100
	// allow a PollTimeout of -1
	if grpConfig.PollTimeout > 0 || grpConfig.PollTimeout == -1 {
		pollTimeout = grpConfig.PollTimeout
	}

	cm := cg.GroupConfig.toConfigMap()

	for i := 0; i < grpConfig.ConsumerCount; i++ {
		workerID := fmt.Sprintf("%s_%d", groupID, i)

		// Each worker gets its own, independent Kafka client connection here — this is
		// the fix for the shared-consumer concurrency bug. Previously, ONE
		// confluentConsumer was created outside this loop and handed to every worker
		// goroutine, meaning ConsumerCount goroutines were all calling
		// Poll/Commit/StoreOffsets/Close concurrently on a single, non-thread-safe
		// client handle — undefined behavior at the librdkafka layer, not just a Go-level
		// data race. Creating one client per worker means ConsumerCount now does what it
		// always looked like it should do: create N genuinely independent members of the
		// same Kafka consumer group, exactly as Kafka's own group-coordination protocol
		// (and the partition-sizing guidance in the README) already assumes.
		confCons := cg.consumerMakeFunc(&cm, cg.GroupConfig.Topics)

		if i == 0 {
			// Preserved for ConsumerHandle's existing public API. With ConsumerCount > 1
			// this now only exposes ONE of several independent consumer handles
			// (arbitrarily the first one created) rather than "the" consumer for this
			// group — ConsumerHandle's contract may be worth revisiting in a follow-up if
			// it's relied on for anything beyond the ConsumerCount == 1 case.
			cg.c = confCons
		}

		cg.Logger.Info("spawning kafka worker", map[string]any{"id": workerID})
		w := worker{
			handler:     handler,
			logger:      cg.Logger,
			consumer:    confCons,
			routeGroup:  cg.GroupConfig.GroupID,
			pollTimeout: pollTimeout,
			killSig:     make(chan struct{}),
			id:          workerID,
		}
		cg.workers[i] = &w
		cg.wg.Add(1)
		go func() {
			// defer, not a trailing call: guarantees wg.Done() still runs even if
			// something inside w.run panics despite the recover() in processMessage
			// (e.g. a panic in worker/orchestration code itself, not just the handler).
			// Without this, an unrecovered panic here would skip wg.Done() entirely and
			// permanently deadlock cg.wg.Wait() for every other worker in the group.
			defer cg.wg.Done()
			w.run(ctx)
		}()
	}

	cg.wg.Wait()
	cg.Logger.Info("kafka worker wait complete")

	// A worker's err is only a *real* failure worth reporting if it's non-nil and isn't
	// one of the two expected, intentional shutdown signals (context cancellation/deadline).
	var workerErrs []error
	for _, w := range cg.workers {
		if w.err != nil && !errors.Is(w.err, context.Canceled) && !errors.Is(w.err, context.DeadlineExceeded) {
			workerErrs = append(workerErrs, fmt.Errorf("%s worker failed with error: %w", w.id, w.err))
		}
	}
	if len(workerErrs) == 0 {
		return ErrCleanShutdown
	}
	return errors.Join(workerErrs...)
}

func (cg *ConsumerGroup) init() {
	var wg sync.WaitGroup
	cg.wg = &wg

	cg.workers = make([]*worker, cg.GroupConfig.ConsumerCount)
	if cg.Logger == nil {
		cg.Logger = logger.NOOP
	}
	if cg.consumerMakeFunc == nil {
		cg.consumerMakeFunc = createConsumer
	}
}
