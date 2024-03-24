package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"sync"

	"github.com/gojekfarm/ziggurat"
)

var ErrCleanShutdown = errors.New("error: clean shutdown of kafka consumers")

type ConsumerGroup struct {
	workers          []*worker
	Logger           ziggurat.StructuredLogger
	GroupConfig      ConsumerConfig
	wg               *sync.WaitGroup
	consumerMakeFunc func(*kafka.ConfigMap, []string) confluentConsumer
}

func (cg *ConsumerGroup) Consume(ctx context.Context, handler ziggurat.Handler) error {
	cg.init()

	consConf := cg.GroupConfig
	groupID := consConf.GroupID
	// sets default pollTimeout of 100ms
	pollTimeout := 100
	// allow a PollTimeout of -1
	if consConf.PollTimeout > 0 || consConf.PollTimeout == -1 {
		pollTimeout = consConf.PollTimeout
	}

	cm := cg.GroupConfig.toConfigMap()
	confCons := cg.consumerMakeFunc(&cm, cg.GroupConfig.Topics)
	for i := 0; i < consConf.ConsumerCount; i++ {
		workerID := fmt.Sprintf("%s_%d", groupID, i)
		w := worker{
			handler:     handler,
			logger:      cg.Logger,
			consumer:    confCons,
			routeGroup:  consConf.RouteGroup,
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
