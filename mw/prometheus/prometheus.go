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
	retrySubsystem   = "retry"
)

const (
	// RouteLabel - Key for route label
	RouteLabel = "route"
	// OutcomeLabel - Key for the handler outcome label (success/panic). Note: ziggurat.Handler.Handle
	// has no error return value, so "error" is not an outcome this middleware can ever observe directly -
	// only whether the handler returned normally or panicked.
	OutcomeLabel = "outcome"
	// QueueLabel - Key for the RabbitMQ queue label on retry/DLQ/replay metrics
	QueueLabel = "queue"
	// ResultLabel - Key for the result label on replay metrics (success/error)
	ResultLabel = "result"
)

const (
	OutcomeSuccess = "success"
	OutcomePanic   = "panic"
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
		Help:      "Events passed on to the handler, partitioned by route and outcome",
	},
	[]string{RouteLabel, OutcomeLabel},
)

// HandlerDurationHistogram - Prometheus histogram for handler duration
var HandlerDurationHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "duration_seconds",
		Help:      "time spent processing events, partitioned by route and outcome",
	},
	[]string{RouteLabel, OutcomeLabel},
)

// RetryPublishedCounter - Prometheus counter for messages republished to a delay/retry queue
var RetryPublishedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "published_total",
		Help:      "Messages republished to a delay/retry queue, partitioned by queue",
	},
	[]string{QueueLabel},
)

// DLQPublishedCounter - Prometheus counter for messages routed to the dead-letter queue
var DLQPublishedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "dlq_published_total",
		Help:      "Messages routed to the dead-letter queue after exhausting retries, partitioned by queue",
	},
	[]string{QueueLabel},
)

// ReplayCounter - Prometheus counter for DLQ replay attempts
var ReplayCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "replay_total",
		Help:      "Messages replayed from the dead-letter queue back to the instant queue, partitioned by queue and result",
	},
	[]string{QueueLabel, ResultLabel},
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
		RetryPublishedCounter,
		DLQPublishedCounter,
		ReplayCounter,
	)
}

// PublishHandlerMetrics - middleware to update registered handler metrics.
//
// A panic inside next.Handle is observed here (recorded as outcome=panic, with duration
// up to the point of the panic) and then deliberately re-panicked, never swallowed: the
// framework's core consume paths no longer recover handler panics (that recovery was
// reverted - see CHANGELOG), and this middleware must not reintroduce swallowing behavior
// as a side effect of adding instrumentation. The only purpose of the recover here is to
// get a chance to record the metric before the panic continues unwinding and crashes the
// process exactly as it would have without this middleware.
func PublishHandlerMetrics(next ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, event *ziggurat.Event) {
		t1 := time.Now()
		outcome := OutcomeSuccess

		defer func() {
			r := recover()
			if r != nil {
				outcome = OutcomePanic
			}

			labels := prometheus.Labels{
				RouteLabel:   event.RoutingPath,
				OutcomeLabel: outcome,
			}
			HandlerDurationHistogram.With(labels).Observe(time.Since(t1).Seconds())
			HandlerEventsCounter.With(labels).Inc()

			if r != nil {
				panic(r) // re-panic: this middleware observes, never swallows
			}
		}()

		next.Handle(ctx, event)
	}
	return ziggurat.HandlerFunc(f)
}
