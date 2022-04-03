package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

var ErrCleanShutdown = errors.New("error: clean shutdown of kafka consumers")

type Streams struct {
	workers      map[string][]*worker
	Logger       ziggurat.StructuredLogger
	StreamConfig StreamConfig
	wg           *sync.WaitGroup
}

func (s *Streams) Stream(ctx context.Context, handler ziggurat.Handler) error {
	groupCount := len(s.StreamConfig)
	var wg sync.WaitGroup
	s.wg = &wg
	s.workers = make(map[string][]*worker, groupCount)
	if s.Logger == nil {
		s.Logger = logger.NOOP
	}
	for _, consConf := range s.StreamConfig {
		groupID := consConf.GroupID
		topics := strings.Split(consConf.Topics, ",")
		confMap := consConf.toConfigMap()
		// sets default pollTimeout of 100ms
		pollTimeout := 100
		// allow a PollTimeout of -1
		if consConf.PollTimeout > 0 || consConf.PollTimeout == -1 {
			pollTimeout = consConf.PollTimeout
		}
		s.workers[groupID] = make([]*worker, consConf.ConsumerCount)
		for i := 0; i < consConf.ConsumerCount; i++ {
			workerID := fmt.Sprintf("%s_%d", groupID, i)
			w := worker{
				handler:     handler,
				logger:      s.Logger,
				topics:      topics,
				confMap:     &confMap,
				routeGroup:  consConf.RouteGroup,
				pollTimeout: pollTimeout,
				killSig:     make(chan struct{}),
				id:          workerID,
			}
			s.workers[groupID][i] = &w
			s.wg.Add(1)
			go func() {
				w.run(ctx)
				s.wg.Done()
			}()
		}
	}
	s.wg.Wait()
	var causes string
	for _, ws := range s.workers {
		for _, w := range ws {
			// okay to use a copy as we are not modifying the worker struct
			if w.err != context.Canceled && w.err != nil {
				err := fmt.Errorf("%s worker failed with error: %w\n", w.id, w.err)
				causes = causes + err.Error()
			}
		}
	}
	if causes == "" {
		return ErrCleanShutdown
	}
	return errors.New(causes)
}
