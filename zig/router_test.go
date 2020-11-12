package zig

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"reflect"
	"sync"
	"testing"
	"time"
)

type routerTestMockApp struct{}

func (r routerTestMockApp) Context() context.Context {
	return context.Background()
}

func (r routerTestMockApp) Router() StreamRouter {
	return nil
}

func (r routerTestMockApp) MessageRetry() MessageRetry {
	return nil
}

func (r routerTestMockApp) Run(router StreamRouter, options RunOptions) chan struct{} {
	return nil
}

func (r routerTestMockApp) Configure(configFunc func(o App) Options) {

}

func (r routerTestMockApp) MetricPublisher() MetricPublisher {
	return nil
}

func (r routerTestMockApp) HTTPServer() HttpServer {
	return nil
}

func (r routerTestMockApp) Config() *Config {
	return &Config{
		StreamRouter: map[string]StreamRouterConfig{
			"foo": {
				InstanceCount:    0,
				BootstrapServers: "baz:9092",
				OriginTopics:     "bar",
				GroupID:          "foo-bar",
			},
		},
	}
}

func (r routerTestMockApp) Stop() {

}

func (r routerTestMockApp) IsRunning() bool {
	return false
}

func TestDefaultRouter_HandlerFunc(t *testing.T) {
	dr := NewRouter()
	topicEntity := "test-entity"
	topicEntityTwo := "test-entity2"
	dr.HandlerFunc(topicEntity, func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})
	dr.HandlerFunc(topicEntityTwo, func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})
	if len(dr.handlerFunctionMap) < 2 {
		t.Errorf("expected %d entries in handlerFunctionMap but got %d", 2, len(dr.handlerFunctionMap))
	}
}

func TestDefaultRouter_StartNoHandlersRegistered(t *testing.T) {
	dr := NewRouter()
	origLogFatal := logFatal
	logFatal = func(err error, msg string, args map[string]interface{}) {
		if err != ErrNoHandlersRegistered {
			t.Errorf("expected error %v, got %v", ErrNoHandlersRegistered, err)
		}
	}
	defer func() {
		logFatal = origLogFatal
	}()
	done, _ := dr.Start(&routerTestMockApp{})
	<-done
}

func TestDefaultRouter_validate(t *testing.T) {
	origLogWarn := logWarn
	expectedArgs := map[string]interface{}{"invalid-entity-name": "baz"}
	logWarn = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(expectedArgs, args) {
			t.Errorf("expected args %v, got %v", expectedArgs, args)
		}
	}
	defer func() {
		logWarn = origLogWarn
	}()
	dr := NewRouter()
	dr.HandlerFunc("baz", func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})
	startConsumers = func(routerCtx context.Context, app App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
		return []*kafka.Consumer{}
	}

	done, _ := dr.Start(&routerTestMockApp{})
	<-done
}

func TestDefaultRouter_Start(t *testing.T) {
	expectedTopicEntity := "foo"
	expectedTopics := []string{"bar"}
	expectedConsumerConfig := &kafka.ConfigMap{
		"bootstrap.servers":        "baz:9092",
		"group.id":                 "foo-bar",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "consumer,broker",
		"enable.auto.offset.store": false,
	}
	dr := NewRouter()
	dr.HandlerFunc("foo", func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})

	startConsumers = func(routerCtx context.Context, app App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
		if !reflect.DeepEqual(expectedConsumerConfig, consumerConfig) {
			t.Errorf("exptected %v but got %v", expectedConsumerConfig, consumerConfig)
		}
		if expectedTopicEntity != topicEntity {
			t.Errorf("expected topic entity %v, got %v", expectedTopicEntity, topicEntity)
		}
		if !reflect.DeepEqual(topics, expectedTopics) {
			t.Errorf("expected %v but got %v", expectedTopics, topics)
		}
		time.Sleep(10 * time.Second)
		return []*kafka.Consumer{}
	}
	done, _ := dr.Start(&routerTestMockApp{})
	<-done
}
