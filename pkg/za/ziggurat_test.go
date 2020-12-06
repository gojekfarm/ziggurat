package za

import (
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"sync/atomic"
	"testing"
	"time"
)

func TestZiggurat_Stop(t *testing.T) {
	app := NewApp()
	kstreams := mock.NewKafkaStreams()
	app.configStore = mock.NewConfigStore()
	app.configValidator = z.ValidatorFunc(func(config *zb.Config) error {
		return nil
	})
	kstreams.StartFunc = func(a z.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(1 * time.Second)
			close(done)
		}()
		return done, nil
	}
	app.streams = kstreams
	handler := z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	<-app.Run(handler, []string{"foo"}, func(opts *RunOptions) {
		opts.Retry = func(c z.ConfigStore) z.MessageRetry {
			return mock.NewRetry()
		}
		opts.MetricPublisher = func(c z.ConfigStore) z.MetricPublisher {
			return mock.NewMetrics()
		}
		opts.HTTPServer = func(c z.ConfigStore) z.Server {
			return mock.NewServer()
		}
	})
}

func TestZiggurat_Run(t *testing.T) {
	startCalled := int32(0)
	stopCalled := int32(0)
	app, kstreams := NewApp(), mock.NewKafkaStreams()
	kstreams.StartFunc = func(a z.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(1 * time.Second)
			close(done)
		}()
		return done, nil
	}
	app.configStore = mock.NewConfigStore()
	app.streams = kstreams
	app.configValidator = z.ValidatorFunc(func(config *zb.Config) error {
		return nil
	})
	handler := z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	<-app.Run(handler, []string{"foo"}, func(opts *RunOptions) {
		opts.MetricPublisher = func(c z.ConfigStore) z.MetricPublisher {
			return mock.NewMetrics()
		}
		opts.Retry = func(c z.ConfigStore) z.MessageRetry {
			return mock.NewRetry()
		}
		opts.HTTPServer = func(c z.ConfigStore) z.Server {
			return mock.NewServer()
		}
		opts.StopCallback = func() {
			atomic.SwapInt32(&startCalled, 1)
		}
		opts.StartCallback = func(a z.App) {
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
	app := NewApp()
	handler := z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	cs := mock.NewConfigStore()
	cs.ConfigFunc = func() *zb.Config {
		return &zb.Config{
			Retry: zb.RetryConfig{
				Enabled: true,
			},
		}
	}
	app.configValidator = z.ValidatorFunc(func(config *zb.Config) error {
		return nil
	})
	app.configStore = cs
	retry := mock.NewRetry()
	metrics := mock.NewMetrics()
	server := mock.NewServer()
	kstreams := mock.NewKafkaStreams()
	server.StartFunc = func(a z.App) error {
		callCount++
		return nil
	}
	retry.StartFunc = func(app z.App) error {
		callCount++
		return nil
	}
	metrics.StartFunc = func(a z.App) error {
		callCount++
		return nil
	}
	kstreams.StartFunc = func(a z.App) (chan struct{}, error) {
		done := make(chan struct{})
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()
		return done, nil
	}
	app.streams = kstreams
	<-app.Run(handler, []string{"foo"}, func(opts *RunOptions) {
		opts.Retry = func(c z.ConfigStore) z.MessageRetry {
			return retry
		}
		opts.MetricPublisher = func(c z.ConfigStore) z.MetricPublisher {
			return metrics
		}
		opts.HTTPServer = func(c z.ConfigStore) z.Server {
			return server
		}
	})

	if callCount < len(app.components()) {
		t.Errorf("expected call count to be %d got %d", len(app.components()), callCount)
	}

}
