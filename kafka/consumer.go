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

	confCons := cg.consumerMakeFunc(&cm, cg.GroupConfig.Topics)

	cg.c = confCons
	for i := 0; i < grpConfig.ConsumerCount; i++ {
		workerID := fmt.Sprintf("%s_%d", groupID, i)
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
			w.run(ctx)
			cg.wg.Done()
		}()
	}

	cg.wg.Wait()
	cg.Logger.Info("kafka worker wait complete")
	var causes string
	for _, w := range cg.workers {
		switch {
		case !errors.Is(w.err, context.Canceled), !errors.Is(w.err, context.DeadlineExceeded):
		case w.err != nil:
			err := fmt.Errorf("%s worker failed with error: %w\n", w.id, w.err)
			causes = causes + err.Error()
		}
	}
	if causes == "" {
		return ErrCleanShutdown
	}
	return errors.New(causes)
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
