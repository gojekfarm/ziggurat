package prometheus

import (
	"context"
	"github.com/gojekfarm/ziggurat/v2"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

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

type ServerOpts func(*http.Server)

func WithAddr(addr string) ServerOpts {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

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
func StartMonitoringServer(ctx context.Context, opts ...ServerOpts) error {
	return startMonitoringServer(ctx, promhttp.Handler(), opts...)
}

// Register - Registers the Prometheus metrics
func Register() {
	prometheus.MustRegister(
		HandlerEventsCounter,
		HandlerDurationHistogram,
	)
}

// PublishHandlerMetrics - middleware to update registered handler metrics
func PublishHandlerMetrics(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, event *ziggurat.Event) {
		t1 := time.Now()
		next.Handle(ctx, event)

		labels := prometheus.Labels{
			RouteLabel: event.RoutingPath,
		}

		HandlerDurationHistogram.With(labels).Observe(time.Since(t1).Seconds())
		HandlerEventsCounter.With(labels).Inc()

	}
	return ziggurat.HandlerFunc(f)
}
