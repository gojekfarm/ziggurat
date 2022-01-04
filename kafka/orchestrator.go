package kafka

//go:generate go run ./librdgen/gen.go

import (
	"context"
	"errors"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"strings"
	"sync"
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
		groupID := consConf.GroupId
		config := NewConsumerConfig(consConf.BootstrapServers, consConf.GroupId)
		topics := strings.Split(consConf.Topics, ",")
		s.workers[groupID] = make([]*worker, consConf.ConsumerCount)
		for i := 0; i < consConf.ConsumerCount; i++ {
			consumer := createConsumer(config, s.Logger, topics)
			workerID := fmt.Sprintf("%s_%d", groupID, i)
			pollTimeout := 1000
			if consConf.PollTimeout > 0 {
				pollTimeout = consConf.PollTimeout
			}
			w := worker{
				handler:     handler,
				logger:      s.Logger,
				consumer:    consumer,
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
			if w.err != context.DeadlineExceeded && w.err != nil {
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
