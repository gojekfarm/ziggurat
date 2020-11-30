package za

import (
	"github.com/gojekfarm/ziggurat-go/pkg/void"
	ztype "github.com/gojekfarm/ziggurat-go/pkg/z"
	"net/http"
)

type RunOptions struct {
	HTTPConfigFunc  func(a ztype.App, h http.Handler)
	StartCallback   func(a ztype.App)
	StopCallback    func()
	HTTPServer      func(c ztype.ConfigStore) ztype.Server
	Retry           func(c ztype.ConfigStore) ztype.MessageRetry
	MetricPublisher func(c ztype.ConfigStore) ztype.MetricPublisher
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
		ro.HTTPConfigFunc = func(a ztype.App, h http.Handler) {}
	}

	if ro.StopCallback == nil {
		ro.StopCallback = func() {}
	}

	if ro.StartCallback == nil {
		ro.StartCallback = func(a ztype.App) {}
	}

}
