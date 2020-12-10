package zapp

import (
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/gojekfarm/ziggurat/zvoid"
	"net/http"
)

type AppOptions struct {
	HTTPConfigFunc  func(a ztype.App, h http.Handler)
	StartCallback   func(a ztype.App)
	StopCallback    func()
	HTTPServer      func(c ztype.ConfigStore) ztype.Server
	Retry           func(c ztype.ConfigStore) ztype.MessageRetry
	MetricPublisher func(c ztype.ConfigStore) ztype.MetricPublisher
}

type Opts = func(opts *AppOptions)

func (ro *AppOptions) setDefaults() {
	if ro.Retry == nil {
		ro.Retry = zvoid.NewRetry
	}
	if ro.MetricPublisher == nil {
		ro.MetricPublisher = zvoid.NewMetrics
	}
	if ro.HTTPServer == nil {
		ro.HTTPServer = zvoid.NewServer
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
