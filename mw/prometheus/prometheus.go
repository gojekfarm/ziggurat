package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"

	"github.com/gojekfarm/ziggurat"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace        = "ziggurat_go"
	handlerSubsystem = "handler"
)

const (
	// RouteLabel - Key for route label
	RouteLabel = "route"
)

// HandlerEventsCounter - Prometheus counter for handled events
var HandlerEventsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "events_total",
		Help:      "Events passed on to the handler, partitioned by route",
	},
	[]string{RouteLabel},
)

// HandlerFailuresCounter - Prometheus counter for handler failures
var HandlerFailuresCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "failures_total",
		Help:      "Event handler failures, partitioned by route",
	},
	[]string{RouteLabel},
)

// HandlerDurationHistogram - Prometheus histogram for handler duration
var HandlerDurationHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "duration_seconds",
		Help:      "time spent processing events, partitioned by route",
	},
	[]string{"route"},
)

// StartMonitoringServer - starts a monitoring server for prometheus
func StartMonitoringServer(ctx context.Context, port string) error {
	return startMonitoringServer(ctx, port, promhttp.Handler())
}

// Register - Registers the Prometheus metrics
func Register() {
	prometheus.MustRegister(
		HandlerEventsCounter,
		HandlerFailuresCounter,
		HandlerDurationHistogram,
	)
}

// PublishHandlerMetrics - middleware to update registered handler metrics
func PublishHandlerMetrics(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, event *ziggurat.Event) error {
		t1 := time.Now()
		err := next.Handle(ctx, event)

		labels := prometheus.Labels{
			RouteLabel: event.Path,
		}

		HandlerDurationHistogram.With(labels).Observe(time.Since(t1).Seconds())

		HandlerEventsCounter.With(labels).Inc()
		if err != nil {
			HandlerFailuresCounter.With(labels).Inc()
		}

		return err
	}
	return ziggurat.HandlerFunc(f)
}
