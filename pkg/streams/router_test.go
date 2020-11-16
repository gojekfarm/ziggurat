package streams

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/cons"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	at "github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"reflect"
	"sync"
	"testing"
)

type routerTestMockApp struct{}

func (r routerTestMockApp) ConfigReader() at.ConfigReader {
	panic("implement me")
}

func (r routerTestMockApp) Context() context.Context {
	return context.Background()
}

func (r routerTestMockApp) Router() at.StreamRouter {
	return nil
}

func (r routerTestMockApp) MessageRetry() at.MessageRetry {
	return nil
}

func (r routerTestMockApp) Run(router at.StreamRouter, options at.RunOptions) chan struct{} {
	return nil
}

func (r routerTestMockApp) Configure(configFunc func(o at.App) at.Options) {

}

func (r routerTestMockApp) MetricPublisher() at.MetricPublisher {
	return nil
}

func (r routerTestMockApp) HTTPServer() at.HttpServer {
	return nil
}

func (r routerTestMockApp) Config() *basic.Config {
	return &basic.Config{
		StreamRouter: map[string]basic.StreamRouterConfig{
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
	dr.HandlerFunc(topicEntity, func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		return at.ProcessingSuccess
	})
	dr.HandlerFunc(topicEntityTwo, func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		return at.ProcessingSuccess
	})
	if len(dr.handlerFunctionMap) < 2 {
		t.Errorf("expected %d entries in handlerFunctionMap but got %d", 2, len(dr.handlerFunctionMap))
	}
}

func TestDefaultRouter_StartNoHandlersRegistered(t *testing.T) {
	dr := NewRouter()
	origLogFatal := logger.LogFatal
	logger.LogFatal = func(err error, msg string, args map[string]interface{}) {
		if err != zerror.ErrNoHandlersRegistered {
			t.Errorf("expected error %v, got %v", zerror.ErrNoHandlersRegistered, err)
		}
	}
	defer func() {
		logger.LogFatal = origLogFatal
	}()
	done, _ := dr.Start(&routerTestMockApp{})
	<-done
}

func TestDefaultRouter_validate(t *testing.T) {
	origLogWarn := logger.LogWarn
	expectedArgs := map[string]interface{}{"invalid-entity-name": "baz"}
	logger.LogWarn = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(expectedArgs, args) {
			t.Errorf("expected args %v, got %v", expectedArgs, args)
		}
	}
	defer func() {
		logger.LogWarn = origLogWarn
	}()
	dr := NewRouter()
	dr.HandlerFunc("baz", func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		return at.ProcessingSuccess
	})
	cons.StartConsumers = func(routerCtx context.Context, app at.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc at.HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
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
	dr.HandlerFunc("foo", func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		return at.ProcessingSuccess
	})

	cons.StartConsumers = func(routerCtx context.Context, app at.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc at.HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
		if !reflect.DeepEqual(expectedConsumerConfig, consumerConfig) {
			t.Errorf("exptected %v but got %v", expectedConsumerConfig, consumerConfig)
		}
		if expectedTopicEntity != topicEntity {
			t.Errorf("expected topic entity %v, got %v", expectedTopicEntity, topicEntity)
		}
		if !reflect.DeepEqual(topics, expectedTopics) {
			t.Errorf("expected %v but got %v", expectedTopics, topics)
		}
		return []*kafka.Consumer{}
	}
	done, _ := dr.Start(&routerTestMockApp{})
	<-done
}
