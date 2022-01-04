package kafka

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	ogCreateConsumer := createConsumer
	ogStoreOffsets := storeOffsets
	ogCloseConsumer := closeConsumer
	ogPollEvent := pollEvent
	defer func() {
		createConsumer = ogCreateConsumer
		storeOffsets = ogStoreOffsets
		closeConsumer = ogCloseConsumer
		pollEvent = ogPollEvent
	}()
	createConsumer = createConsumerMock
	closeConsumer = closeConsumerMock
	pollEvent = pollEventMock
	storeOffsets = storeOffsetsMock

	c := createConsumer(&kafka.ConfigMap{}, nil, []string{})
	called := false
	h := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
		called = true
		return nil
	})
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()
	w := worker{
		handler:     h,
		logger:      logger.NOOP,
		consumer:    c,
		routeGroup:  "",
		pollTimeout: 100,
		id:          "1",
		killSig:     make(chan struct{}),
	}
	w.run(ctx)
	if !called {
		t.Error("expected called to be true")
	}
}

func TestWorker_Kill(t *testing.T) {
	ogCreateConsumer := createConsumer
	ogStoreOffsets := storeOffsets
	ogCloseConsumer := closeConsumer
	ogPollEvent := pollEvent
	defer func() {
		createConsumer = ogCreateConsumer
		storeOffsets = ogStoreOffsets
		closeConsumer = ogCloseConsumer
		pollEvent = ogPollEvent
	}()
	createConsumer = createConsumerMock
	closeConsumer = closeConsumerMock
	pollEvent = pollEventMock
	storeOffsets = storeOffsetsMock

	c := createConsumer(&kafka.ConfigMap{}, nil, []string{})

	w := worker{
		handler: ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
			return nil
		}),
		logger:      logger.NOOP,
		consumer:    c,
		routeGroup:  "",
		pollTimeout: 100,
		id:          "1",
		killSig:     make(chan struct{}),
	}

	wait := make(chan struct{})
	go func() {
		w.run(context.Background())
		wait <- struct{}{}
	}()
	time.AfterFunc(100*time.Millisecond, func() {
		w.kill()
	})

	<-wait
	killErr := ErrorWorkerKilled{workerID: "1"}
	if e := w.err; !errors.Is(e, killErr) {
		t.Errorf("expected error to be [%v] got [%v]", killErr, e)
	}
}
