package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/mock"
)

func tp(topic string, p int32) kafka.TopicPartition {
	return kafka.TopicPartition{Topic: &topic, Partition: p}
}

// ---------------- per-partition pause tracking ----------------

func TestPauseTracker_OnlyRegainedPartitionsReportAPause(t *testing.T) {
	tr := newRebalancePauseTracker()
	t0 := time.Now()
	tr.revoked("c", []kafka.TopicPartition{tp("a", 0), tp("a", 1)}, t0)

	// partition 5 was never revoked (a partition this member kept or just received)
	pauses := tr.assigned("c", []kafka.TopicPartition{tp("a", 1), tp("a", 5)}, t0.Add(2*time.Second))
	if len(pauses) != 1 || pauses[0] != 2*time.Second {
		t.Fatalf("expected exactly one 2s pause for a/1, got %v", pauses)
	}
	// a/0 is still pending, a/1 was consumed
	if pauses := tr.assigned("c", []kafka.TopicPartition{tp("a", 1)}, t0.Add(3*time.Second)); len(pauses) != 0 {
		t.Fatalf("a/1 must be forgotten after it is reported, got %v", pauses)
	}
	if pauses := tr.assigned("c", []kafka.TopicPartition{tp("a", 0)}, t0.Add(4*time.Second)); len(pauses) != 1 {
		t.Fatalf("a/0 must still be tracked, got %v", pauses)
	}
}

func TestPauseTracker_StaleRevokeIsNotReportedAsAPause(t *testing.T) {
	// The bug this guards: under cooperative rebalancing a member can revoke a partition
	// and regain it hours later in an unrelated rebalance. That must not be reported as
	// an hours-long "pause".
	tr := newRebalancePauseTracker()
	t0 := time.Now()
	tr.revoked("c", []kafka.TopicPartition{tp("a", 0)}, t0)
	if pauses := tr.assigned("c", []kafka.TopicPartition{tp("a", 0)}, t0.Add(pauseTrackerTTL+time.Second)); len(pauses) != 0 {
		t.Fatalf("expected a stale revoke to be dropped, got pauses %v", pauses)
	}
}

func TestPauseTracker_RevokePrunesStaleEntries(t *testing.T) {
	tr := newRebalancePauseTracker()
	t0 := time.Now()
	tr.revoked("c", []kafka.TopicPartition{tp("never-returns", 0)}, t0)
	tr.revoked("c", []kafka.TopicPartition{tp("a", 0)}, t0.Add(pauseTrackerTTL+time.Minute))
	if n := len(tr.revokedAt); n != 1 {
		t.Fatalf("expected the stale entry to be pruned (1 entry left), got %d", n)
	}
}

func TestPauseTracker_IsolatedPerConsumer(t *testing.T) {
	tr := newRebalancePauseTracker()
	t0 := time.Now()
	tr.revoked("c1", []kafka.TopicPartition{tp("a", 0)}, t0)
	if pauses := tr.assigned("c2", []kafka.TopicPartition{tp("a", 0)}, t0.Add(time.Second)); len(pauses) != 0 {
		t.Fatalf("a revoke on c1 must not produce a pause on c2, got %v", pauses)
	}
}

// ---------------- rebalance metrics ----------------

func TestHandleRebalance_EmitsMetrics(t *testing.T) {
	const group = "obs-rebalance-group"
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	tracker := newRebalancePauseTracker()
	parts := []kafka.TopicPartition{tp("obs-topic", 0)}
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("IncrementalUnassign", mock.Anything).Return(nil)
	mc.On("IncrementalAssign", mock.Anything).Return(nil)

	if err := handleRebalance(log, tracker, group, "k", "m", mc, kafka.RevokedPartitions{Partitions: parts}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := handleRebalance(log, tracker, group, "k", "m", mc, kafka.AssignedPartitions{Partitions: parts}); err != nil {
		t.Fatal(err)
	}

	if got := testutil.ToFloat64(prometheus.RebalanceCounter.WithLabelValues(group, "cooperative", "revoked")); got != 1 {
		t.Errorf("revoked counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.RebalanceCounter.WithLabelValues(group, "cooperative", "assigned")); got != 1 {
		t.Errorf("assigned counter = %v, want 1", got)
	}
	if n := testutil.CollectAndCount(prometheus.RebalancePauseHistogram); n == 0 {
		t.Error("expected a rebalance_pause_seconds observation for the regained partition")
	}
	fields := log.infoFields[len(log.infoFields)-1]
	if fields["group"] != group || fields["protocol"] != "cooperative" {
		t.Errorf("expected group/protocol in log fields, got %v", fields)
	}
}

func TestHandleRebalance_EagerLabelsProtocol(t *testing.T) {
	const group = "obs-eager-group"
	mc := &rebalanceMockConsumer{protocol: "EAGER"}
	mc.On("Assign", mock.Anything).Return(nil)
	if err := handleRebalance(&rebalanceLogCapture{}, newRebalancePauseTracker(), group, "k", "m", mc, kafka.AssignedPartitions{Partitions: []kafka.TopicPartition{tp("t", 0)}}); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(prometheus.RebalanceCounter.WithLabelValues(group, "eager", "assigned")); got != 1 {
		t.Errorf("eager assigned counter = %v, want 1", got)
	}
}

func TestHandleRebalance_RevokeCommitErrorKinds(t *testing.T) {
	const group = "obs-commit-group"
	mc := &rebalanceMockConsumer{}
	mc.On("IncrementalUnassign", mock.Anything).Return(nil)
	mc.On("Commit").Return([]kafka.TopicPartition{}, kafka.NewError(kafka.ErrNoOffset, "none", false)).Once()
	mc.On("Commit").Return([]kafka.TopicPartition{}, errors.New("coordinator down")).Once()

	revoke := kafka.RevokedPartitions{Partitions: []kafka.TopicPartition{tp("t", 0)}}
	for i := 0; i < 2; i++ {
		if err := handleRebalance(&rebalanceLogCapture{}, newRebalancePauseTracker(), group, "k", "m", mc, revoke); err != nil {
			t.Fatal(err)
		}
	}
	if got := testutil.ToFloat64(prometheus.RevokeCommitErrorsCounter.WithLabelValues(group, "no_offset")); got != 1 {
		t.Errorf("no_offset = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.RevokeCommitErrorsCounter.WithLabelValues(group, "error")); got != 1 {
		t.Errorf("error = %v, want 1", got)
	}
}

// ---------------- lag from librdkafka statistics ----------------

const statsTwoPartitions = `{"topics":{"orders":{"partitions":{
  "0":{"partition":0,"consumer_lag":42},
  "1":{"partition":1,"consumer_lag":-1},
  "-1":{"partition":-1,"consumer_lag":-1}}}}}`

func TestLagReporter_ReportsOwnedPartitionsOnly(t *testing.T) {
	r := newLagReporter("lag-g1")
	if err := r.update(statsTwoPartitions); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(prometheus.ConsumerLagGauge.WithLabelValues("lag-g1", "orders", "0")); got != 42 {
		t.Errorf("lag for orders/0 = %v, want 42", got)
	}
	// 1 (lag -1) and -1 (internal) must not create series.
	if n := testutil.CollectAndCount(prometheus.ConsumerLagGauge); n != 1 {
		t.Errorf("expected exactly 1 lag series, got %d", n)
	}
	r.clear()
}

func TestLagReporter_RemovesSeriesForPartitionsNoLongerOwned(t *testing.T) {
	r := newLagReporter("lag-g2")
	if err := r.update(statsTwoPartitions); err != nil {
		t.Fatal(err)
	}
	// next stats: orders/0 was rebalanced away, orders/2 arrived
	if err := r.update(`{"topics":{"orders":{"partitions":{"2":{"partition":2,"consumer_lag":7}}}}}`); err != nil {
		t.Fatal(err)
	}
	if n := testutil.CollectAndCount(prometheus.ConsumerLagGauge); n != 1 {
		t.Fatalf("expected only orders/2 to remain, got %d series", n)
	}
	if got := testutil.ToFloat64(prometheus.ConsumerLagGauge.WithLabelValues("lag-g2", "orders", "2")); got != 7 {
		t.Errorf("lag for orders/2 = %v, want 7", got)
	}
	r.clear()
	if n := testutil.CollectAndCount(prometheus.ConsumerLagGauge); n != 0 {
		t.Errorf("clear must remove every series, %d left", n)
	}
}

func TestLagReporter_BadPayload(t *testing.T) {
	r := newLagReporter("lag-g3")
	if err := r.update(""); err == nil {
		t.Error("expected error for empty payload")
	}
	if err := r.update("{not json"); err == nil {
		t.Error("expected error for malformed payload")
	}
}

// ---------------- config ----------------

func TestToConfigMap_StatisticsInterval(t *testing.T) {
	base := ConsumerConfig{BootstrapServers: "localhost:9092", GroupID: "g"}

	assertConfigValue(t, base.toConfigMap(), "statistics.interval.ms", DefaultStatisticsIntervalMS)

	base.StatisticsIntervalMS = 5000
	assertConfigValue(t, base.toConfigMap(), "statistics.interval.ms", 5000)

	base.StatisticsIntervalMS = -1
	if _, ok := base.toConfigMap()["statistics.interval.ms"]; ok {
		t.Error("a negative StatisticsIntervalMS must leave statistics disabled")
	}
}

func TestToConfigMap_DebugIsOffByDefault(t *testing.T) {
	cm := ConsumerConfig{BootstrapServers: "localhost:9092", GroupID: "g"}.toConfigMap()
	if v, ok := cm["debug"]; ok {
		t.Errorf("expected no librdkafka debug contexts by default, got %v", v)
	}
	cm = ConsumerConfig{BootstrapServers: "localhost:9092", GroupID: "g", DebugLevel: "cgrp"}.toConfigMap()
	assertConfigValue(t, cm, "debug", "cgrp")
}

// ---------------- worker metrics ----------------

func TestExitReason(t *testing.T) {
	cases := map[string]error{
		prometheus.WorkerExitContextDone: context.Canceled,
		prometheus.WorkerExitKilled:      ErrorWorkerKilled{workerID: "w"},
		prometheus.WorkerExitFatalError:  errors.New("boom"),
	}
	for want, err := range cases {
		if got := exitReason(err); got != want {
			t.Errorf("exitReason(%v) = %q, want %q", err, got, want)
		}
	}
	if got := exitReason(nil); got != prometheus.WorkerExitContextDone {
		t.Errorf("exitReason(nil) = %q", got)
	}
}

func TestWorker_RecordsErrorsAndLifecycleMetrics(t *testing.T) {
	const group = "obs-worker-group"
	mc := &MockConsumer{}
	w := worker{
		handler:     ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {}),
		logger:      logger.NOOP,
		consumer:    mc,
		routeGroup:  group,
		pollTimeout: 10,
		killSig:     make(chan struct{}),
		id:          "obs-worker",
	}
	mc.On("Logs").Return(make(chan kafka.LogEvent))
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("Close").Return(nil)
	nonFatal := kafka.NewError(kafka.ErrTransport, "transport", false)
	fatal := kafka.NewError(kafka.ErrFatal, "fatal", true)
	mc.On("Poll", 10).Return(nonFatal).Once()
	mc.On("Poll", 10).Return(fatal).Once()

	w.run(context.Background())

	if got := testutil.ToFloat64(prometheus.KafkaErrorsCounter.WithLabelValues(group, kafka.ErrTransport.String(), "false")); got != 1 {
		t.Errorf("non-fatal error counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.KafkaErrorsCounter.WithLabelValues(group, kafka.ErrFatal.String(), "true")); got != 1 {
		t.Errorf("fatal error counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.WorkerExitsCounter.WithLabelValues(group, prometheus.WorkerExitFatalError)); got != 1 {
		t.Errorf("worker exit (fatal_error) = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.WorkersRunningGauge.WithLabelValues(group)); got != 0 {
		t.Errorf("workers_running must return to 0 after the worker exits, got %v", got)
	}
}
