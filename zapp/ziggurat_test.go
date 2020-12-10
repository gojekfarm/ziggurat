package zapp

import (
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"sync/atomic"
	"testing"
	"time"
)

func TestZiggurat_Stop(t *testing.T) {
	app := New()
	callCount := 0
	kstreams := mock.NewKafkaStreams()
	app.configStore = mock.NewConfigStore()
	retry := mock.NewRetry()
	metrics := mock.NewMetrics()
	server := mock.NewServer()
	retry.StopFunc = func(app ztype.App) {
		callCount++
	}
	metrics.StopFunc = func(app ztype.App) {
		callCount++
	}
	server.StopFunc = func(a ztype.App) {
		callCount++
	}
	app.configValidator = ztype.ValidatorFunc(func(config *zbase.Config) error {
		return nil
	})
	kstreams.StartFunc = func(a ztype.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(1 * time.Second)
			close(done)
		}()
		return done, nil
	}
	app.streams = kstreams
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	go func() {
		time.Sleep(100 * time.Millisecond)
		app.Stop()
	}()
	<-app.Run(handler, []string{"foo"}, func(opts *AppOptions) {
		opts.Retry = func(c ztype.ConfigStore) ztype.MessageRetry {
			return retry
		}
		opts.MetricPublisher = func(c ztype.ConfigStore) ztype.MetricPublisher {
			return metrics
		}
		opts.HTTPServer = func(c ztype.ConfigStore) ztype.Server {
			return server
		}
	})
	if callCount < len(app.components()) {
		t.Errorf("expected call count to be %d got %d", callCount, app.components())
	}
}

func TestZiggurat_Run(t *testing.T) {
	startCalled := int32(0)
	stopCalled := int32(0)
	app, kstreams := New(), mock.NewKafkaStreams()
	kstreams.StartFunc = func(a ztype.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(1 * time.Second)
			close(done)
		}()
		return done, nil
	}
	app.configStore = mock.NewConfigStore()
	app.streams = kstreams
	app.configValidator = ztype.ValidatorFunc(func(config *zbase.Config) error {
		return nil
	})
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	<-app.Run(handler, []string{"foo"}, func(opts *AppOptions) {
		opts.MetricPublisher = func(c ztype.ConfigStore) ztype.MetricPublisher {
			return mock.NewMetrics()
		}
		opts.Retry = func(c ztype.ConfigStore) ztype.MessageRetry {
			return mock.NewRetry()
		}
		opts.HTTPServer = func(c ztype.ConfigStore) ztype.Server {
			return mock.NewServer()
		}
		opts.StopCallback = func() {
			atomic.SwapInt32(&startCalled, 1)
		}
		opts.StartCallback = func(a ztype.App) {
			atomic.SwapInt32(&stopCalled, 1)
		}
	})
	if atomic.LoadInt32(&startCalled) != 1 {
		t.Errorf("start was not called")
	}
	if atomic.LoadInt32(&stopCalled) != 1 {
		t.Errorf("stop was not called")
	}
}

func TestZiggurat_start(t *testing.T) {
	callCount := 0
	app := New()
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	cs := mock.NewConfigStore()
	cs.ConfigFunc = func() *zbase.Config {
		return &zbase.Config{
			Retry: zbase.RetryConfig{
				Enabled: true,
			},
		}
	}
	app.configValidator = ztype.ValidatorFunc(func(config *zbase.Config) error {
		return nil
	})
	app.configStore = cs
	retry := mock.NewRetry()
	metrics := mock.NewMetrics()
	server := mock.NewServer()
	kstreams := mock.NewKafkaStreams()
	server.StartFunc = func(a ztype.App) error {
		callCount++
		return nil
	}
	retry.StartFunc = func(app ztype.App) error {
		callCount++
		return nil
	}
	metrics.StartFunc = func(a ztype.App) error {
		callCount++
		return nil
	}
	kstreams.StartFunc = func(a ztype.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()
		return done, nil
	}
	app.streams = kstreams
	<-app.Run(handler, []string{"foo"}, func(opts *AppOptions) {
		opts.Retry = func(c ztype.ConfigStore) ztype.MessageRetry {
			return retry
		}
		opts.MetricPublisher = func(c ztype.ConfigStore) ztype.MetricPublisher {
			return metrics
		}
		opts.HTTPServer = func(c ztype.ConfigStore) ztype.Server {
			return server
		}
	})

	if callCount < len(app.components()) {
		t.Errorf("expected call count to be %d got %d", len(app.components()), callCount)
	}
}

func TestZiggurat_RunWithOptions(t *testing.T) {
	app := New()
	cs := mock.NewConfigStore()
	app.configValidator = ztype.ValidatorFunc(func(config *zbase.Config) error {
		return nil
	})
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	app.configStore = cs
	app.Run(handler, []string{"foo"}, func(opts *AppOptions) {
		opts.HTTPServer = nil
		opts.MetricPublisher = nil
		opts.Retry = nil
		opts.StartCallback = func(a ztype.App) {
			if a.HTTPServer() == nil || a.MetricPublisher() == nil || a.MessageRetry() == nil {
				t.Errorf("expected components to be not nil but are nil")
			}
		}
		opts.StopCallback = nil
	})
}
