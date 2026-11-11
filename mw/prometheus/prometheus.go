package prometheus

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/internal/eventmeta"
	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace        = "ziggurat_go"
	handlerSubsystem = "handler"
	retrySubsystem   = "retry"
	kafkaSubsystem   = "kafka"
)

const (
	// RouteLabel - Key for route label. By default it is the full routing path
	// (<group>/<topic>/<partition> for Kafka); see WithoutPartitionInRoute.
	RouteLabel = "route"
	// OutcomeLabel - Key for the handler outcome label (success/panic). Note: ziggurat.Handler.Handle
	// has no error return value, so "error" is not an outcome this middleware can ever observe directly -
	// only whether the handler returned normally or panicked.
	OutcomeLabel = "outcome"
	// TopicLabel - Key for the Kafka topic label (empty for events that carry no topic)
	TopicLabel = "topic"
	// GroupLabel - Key for the Kafka consumer group label (empty for events that carry no group)
	GroupLabel = "group"
	// PartitionLabel - Key for the Kafka partition label (only used by the lag gauge)
	PartitionLabel = "partition"
	// QueueLabel - Key for the RabbitMQ queue label on retry/DLQ/replay metrics
	QueueLabel = "queue"
	// ResultLabel - Key for the result label on replay / publish metrics (success/error)
	ResultLabel = "result"
	// ReasonLabel - Key for the publish error reason label (nack/marshal/publish)
	ReasonLabel = "reason"
	// ProtocolLabel - Key for the Kafka rebalance protocol label (cooperative/eager)
	ProtocolLabel = "protocol"
	// EventLabel - Key for the Kafka rebalance event label (assigned/revoked)
	EventLabel = "event"
	// KindLabel - Key for the revoke-commit error kind label (error/no_offset)
	KindLabel = "kind"
	// CodeLabel - Key for the Kafka error code label (librdkafka error code name)
	CodeLabel = "code"
	// FatalLabel - Key for the Kafka error fatal label (true/false)
	FatalLabel = "fatal"
)

const (
	OutcomeSuccess = "success"
	OutcomePanic   = "panic"
)

// Values for the ReasonLabel on WorkerExitsCounter.
const (
	WorkerExitContextDone = "context_done"
	WorkerExitKilled      = "killed"
	WorkerExitFatalError  = "fatal_error"
)

// Values for the ReasonLabel on RetryPublishErrorsCounter.
const (
	PublishErrNack    = "nack"
	PublishErrMarshal = "marshal"
	PublishErrPublish = "publish"
)

type ServerOpts func(*http.Server)

// WithAddr sets the address the monitoring server listens on. The default is
// "localhost:9090", which is only reachable from inside the pod/host: set this (e.g.
// ":9090") for an external scraper to reach /metrics.
func WithAddr(addr string) ServerOpts {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

// defaultHandlerDurationBuckets spans 5ms to ~41s (5ms * 2^n). The client_golang default
// stops at 10s, which hides slow-handler tails; these keep a usable p99 up to ~40s.
var defaultHandlerDurationBuckets = prometheus.ExponentialBuckets(0.005, 2, 14)

func newHandlerDurationHistogram(buckets []float64) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: handlerSubsystem,
			Name:      "duration_seconds",
			Help:      "time spent processing events, partitioned by route and outcome",
			Buckets:   buckets,
		},
		handlerLabels,
	)
}

var handlerLabels = []string{RouteLabel, OutcomeLabel, TopicLabel, GroupLabel}

// HandlerEventsCounter - Prometheus counter for handled events
var HandlerEventsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "events_total",
		Help:      "Events passed on to the handler, partitioned by route, outcome, topic and group",
	},
	handlerLabels,
)

// HandlerDurationHistogram - Prometheus histogram for handler duration. To change the
// buckets call SetHandlerDurationBuckets once at start-up, before Register.
var HandlerDurationHistogram = newHandlerDurationHistogram(defaultHandlerDurationBuckets)

// SetHandlerDurationBuckets replaces HandlerDurationHistogram with one that uses the
// given buckets. Use it to put a bucket boundary exactly on your latency SLO threshold,
// otherwise a "fraction of events under T" SLI is only an interpolation.
//
// Call it once at start-up, before Register and before any handler runs. It is not safe
// for concurrent use with PublishHandlerMetrics.
func SetHandlerDurationBuckets(buckets []float64) {
	HandlerDurationHistogram = newHandlerDurationHistogram(buckets)
}

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

// RetryPublishErrorsCounter - Prometheus counter for failed retry/DLQ/publish calls. With
// rabbitmq.WithPublisherConfirms enabled a broker nack shows up here as reason=nack.
var RetryPublishErrorsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "publish_errors_total",
		Help:      "Failed RabbitMQ retry/DLQ publishes, partitioned by queue and reason (nack/marshal/publish)",
	},
	[]string{QueueLabel, ReasonLabel},
)

// RetryPublishDurationHistogram - Prometheus histogram for RabbitMQ publish latency. With
// publisher confirms enabled this includes the wait for the broker acknowledgement.
var RetryPublishDurationHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "publish_duration_seconds",
		Help:      "Time spent publishing a message to RabbitMQ (includes confirm wait when enabled)",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	},
	[]string{QueueLabel, ResultLabel},
)

// RetryPubpoolInUseGauge - Prometheus gauge for RabbitMQ publishers currently checked out
var RetryPubpoolInUseGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: retrySubsystem,
		Name:      "publisher_pool_in_use",
		Help:      "RabbitMQ publishers currently checked out of the publisher pool",
	},
)

// RebalanceCounter - Prometheus counter for Kafka rebalance callbacks
var RebalanceCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "rebalance_total",
		Help:      "Kafka rebalance callbacks, partitioned by group, protocol (cooperative/eager) and event (assigned/revoked)",
	},
	[]string{GroupLabel, ProtocolLabel, EventLabel},
)

// RebalancePauseHistogram - Prometheus histogram for how long a partition was unowned
// by this member between being revoked and being assigned back. One observation per
// regained partition; stale entries (see kafka's pause tracker) are not observed.
var RebalancePauseHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "rebalance_pause_seconds",
		Help:      "Time a partition was unowned between revoke and re-assign on the same member",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	},
	[]string{GroupLabel, ProtocolLabel},
)

// RevokeCommitErrorsCounter - Prometheus counter for offset commits that fail or find
// nothing to commit when partitions are revoked
var RevokeCommitErrorsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "revoke_commit_errors_total",
		Help:      "Synchronous commits on partition revocation that failed (kind=error) or had no offsets (kind=no_offset)",
	},
	[]string{GroupLabel, KindLabel},
)

// KafkaErrorsCounter - Prometheus counter for errors returned by librdkafka's poll loop
var KafkaErrorsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "errors_total",
		Help:      "Errors surfaced by the Kafka consumer poll loop, partitioned by group, librdkafka error code and fatal",
	},
	[]string{GroupLabel, CodeLabel, FatalLabel},
)

// WorkerExitsCounter - Prometheus counter for Kafka worker goroutines that exited
var WorkerExitsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "worker_exits_total",
		Help:      "Kafka workers that exited, partitioned by group and reason (context_done/killed/fatal_error)",
	},
	[]string{GroupLabel, ReasonLabel},
)

// WorkersRunningGauge - Prometheus gauge for Kafka workers currently running
var WorkersRunningGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "workers_running",
		Help:      "Kafka workers currently running, partitioned by group. Compare with ConsumerCount",
	},
	[]string{GroupLabel},
)

// ConsumerLagGauge - Prometheus gauge for consumer lag (messages), fed by librdkafka's
// statistics. Series for partitions this member no longer owns are removed.
var ConsumerLagGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: kafkaSubsystem,
		Name:      "consumer_lag",
		Help:      "Messages between the partition high watermark and this group's stored offset, per owned partition",
	},
	[]string{GroupLabel, TopicLabel, PartitionLabel},
)

// AppInfoGauge - constant 1 gauge carrying the application version, so dashboards and
// alerts can split canary from baseline: `metric * on() group_left(version) ziggurat_go_app_info`.
var AppInfoGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "app_info",
		Help:      "Application build information; value is always 1",
	},
	[]string{"version"},
)

// SetAppVersion records the application's version in ziggurat_go_app_info. Call it once
// at start-up with your own release/commit identifier.
func SetAppVersion(version string) {
	AppInfoGauge.Reset()
	AppInfoGauge.WithLabelValues(version).Set(1)
}

func collectors() []prometheus.Collector {
	return []prometheus.Collector{
		HandlerEventsCounter,
		HandlerDurationHistogram,
		RetryPublishedCounter,
		DLQPublishedCounter,
		ReplayCounter,
		RetryPublishErrorsCounter,
		RetryPublishDurationHistogram,
		RetryPubpoolInUseGauge,
		RebalanceCounter,
		RebalancePauseHistogram,
		RevokeCommitErrorsCounter,
		KafkaErrorsCounter,
		WorkerExitsCounter,
		WorkersRunningGauge,
		ConsumerLagGauge,
		AppInfoGauge,
	}
}

// metricsHandler serves /metrics with OpenMetrics enabled, which is required for the
// trace exemplars on the handler duration histogram to be exposed. Scrapers that don't
// ask for OpenMetrics still get the classic text format.
func metricsHandler() http.Handler {
	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

// StartMonitoringServer - starts a monitoring server for prometheus. It listens on
// localhost:9090 unless WithAddr is given, and every path serves the metrics.
func StartMonitoringServer(ctx context.Context, opts ...ServerOpts) error {
	return startMonitoringServer(ctx, metricsHandler(), opts...)
}

// Register - Registers the Prometheus metrics with the default registerer. It is safe
// to call more than once; metrics that are already registered are skipped.
func Register() {
	for _, c := range collectors() {
		if err := prometheus.Register(c); err != nil {
			var are prometheus.AlreadyRegisteredError
			if errors.As(err, &are) {
				continue
			}
			panic(err)
		}
	}
}

type handlerMetricsConfig struct {
	withoutPartition bool
}

// HandlerMetricsOption configures HandlerMetrics.
type HandlerMetricsOption func(*handlerMetricsConfig)

// WithoutPartitionInRoute drops the partition from the `route` label of Kafka events
// (`<group>/<topic>/<partition>` becomes `<group>/<topic>`). Partition is the highest
// cardinality dimension of the handler metrics (series = partitions x outcomes x
// histogram buckets); use this unless you really slice by partition. Events that do not
// carry Kafka topic/partition metadata keep their full routing path.
func WithoutPartitionInRoute() HandlerMetricsOption {
	return func(c *handlerMetricsConfig) { c.withoutPartition = true }
}

// routeInfo returns the route, topic and group label values for an event. topic and
// group are only known for events carrying Kafka topic and partition metadata whose
// routing path ends in "/<topic>/<partition>" (what kafka.ConsumerGroup produces).
func routeInfo(event *ziggurat.Event, withoutPartition bool) (route, topic, group string) {
	group, topic, _, ok := eventmeta.Kafka(event)
	if !ok {
		return event.RoutingPath, "", ""
	}
	if withoutPartition {
		return group + "/" + topic, topic, group
	}
	return event.RoutingPath, topic, group
}

// PublishHandlerMetrics - middleware to update registered handler metrics, using the
// full routing path as the route label. See HandlerMetrics for options.
//
// A panic inside next.Handle is observed here (recorded as outcome=panic, with duration
// up to the point of the panic) and then deliberately re-panicked, never swallowed: the
// framework's core consume paths no longer recover handler panics (that recovery was
// reverted - see CHANGELOG), and this middleware must not reintroduce swallowing behavior
// as a side effect of adding instrumentation. The only purpose of the recover here is to
// get a chance to record the metric before the panic continues unwinding and crashes the
// process exactly as it would have without this middleware.
func PublishHandlerMetrics(next ziggurat.Handler) ziggurat.Handler {
	return HandlerMetrics()(next)
}

// HandlerMetrics returns the handler metrics middleware, configured with opts. It has
// the same panic contract as PublishHandlerMetrics. When the handler context carries a
// sampled trace, the duration observation gets a `trace_id` exemplar so a slow bucket
// links straight to a trace (needs OpenMetrics scraping and Prometheus exemplar storage).
func HandlerMetrics(opts ...HandlerMetricsOption) func(ziggurat.Handler) ziggurat.Handler {
	var cfg handlerMetricsConfig
	for _, o := range opts {
		o(&cfg)
	}
	return func(next ziggurat.Handler) ziggurat.Handler {
		f := func(ctx context.Context, event *ziggurat.Event) {
			t1 := time.Now()
			outcome := OutcomeSuccess

			defer func() {
				r := recover()
				if r != nil {
					outcome = OutcomePanic
				}

				route, topic, group := routeInfo(event, cfg.withoutPartition)
				labels := prometheus.Labels{
					RouteLabel:   route,
					OutcomeLabel: outcome,
					TopicLabel:   topic,
					GroupLabel:   group,
				}
				observeDuration(ctx, labels, time.Since(t1).Seconds())
				HandlerEventsCounter.With(labels).Inc()

				if r != nil {
					panic(r) // re-panic: this middleware observes, never swallows
				}
			}()

			next.Handle(ctx, event)
		}
		return ziggurat.HandlerFunc(f)
	}
}

func observeDuration(ctx context.Context, labels prometheus.Labels, seconds float64) {
	obs := HandlerDurationHistogram.With(labels)
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() && sc.IsSampled() {
		if eo, ok := obs.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(seconds, prometheus.Labels{"trace_id": sc.TraceID().String()})
			return
		}
	}
	obs.Observe(seconds)
}
