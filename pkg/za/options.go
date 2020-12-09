package za

import (
	"github.com/gojekfarm/ziggurat/pkg/void"
	z "github.com/gojekfarm/ziggurat/pkg/z"
	"net/http"
)

type RunOptions struct {
	HTTPConfigFunc  func(a z.App, h http.Handler)
	StartCallback   func(a z.App)
	StopCallback    func()
	HTTPServer      func(c z.ConfigStore) z.Server
	Retry           func(c z.ConfigStore) z.MessageRetry
	MetricPublisher func(c z.ConfigStore) z.MetricPublisher
}

type Opts = func(opts *RunOptions)

func (ro *RunOptions) setDefaults() {
	if ro.Retry == nil {
		ro.Retry = void.NewRetry
	}
	if ro.MetricPublisher == nil {
		ro.MetricPublisher = void.NewMetrics
	}
	if ro.HTTPServer == nil {
		ro.HTTPServer = void.NewServer
	}

	if ro.HTTPConfigFunc == nil {
		ro.HTTPConfigFunc = func(a z.App, h http.Handler) {}
	}

	if ro.StopCallback == nil {
		ro.StopCallback = func() {}
	}

	if ro.StartCallback == nil {
		ro.StartCallback = func(a z.App) {}
	}

}
