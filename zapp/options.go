package zapp

import (
	"github.com/gojekfarm/ziggurat/ztype"
)

type ZigOptions func(z *Ziggurat)

func WithRetry(retryComponent ztype.MessageRetry) ZigOptions {
	return func(z *Ziggurat) {
		z.messageRetry = retryComponent
	}
}

func WithHTTPServer(server ztype.Server) ZigOptions {
	return func(z *Ziggurat) {
		z.httpServer = server
	}
}

func WithMetrics(metricsComponent ztype.MetricPublisher) ZigOptions {
	return func(z *Ziggurat) {
		z.metricPublisher = metricsComponent
	}
}

func WithLogLevel(level string) ZigOptions {
	return func(z *Ziggurat) {
		z.logLevel = level
	}
}
