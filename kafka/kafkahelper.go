package kafka

import (
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"
)

// rebalanceLogger is the subset of ziggurat.StructuredLogger that the rebalance
// callback needs. Kept as a local interface so the kafka package does not need to
// import the top-level ziggurat package just for its logger interface, avoiding a
// circular dependency.
type rebalanceLogger interface {
	Error(string, error, ...map[string]interface{})
	Info(string, ...map[string]interface{})
}

// pauseTrackerTTL bounds how long a revoked partition is remembered. A partition that
// comes back to the same member later than this is a different rebalance, not the
// "pause" of the one that revoked it (with cooperative rebalancing a member can revoke a
// partition and only regain it hours later), so it is neither reported nor kept.
const pauseTrackerTTL = 5 * time.Minute

// pauseKey identifies one partition on one consumer instance.
type pauseKey struct {
	consumer  any
	topic     string
	partition int32
}

func newPauseKey(consumer any, tp kafka.TopicPartition) pauseKey {
	k := pauseKey{consumer: consumer, partition: tp.Partition}
	if tp.Topic != nil {
		k.topic = *tp.Topic
	}
	return k
}

// rebalancePauseTracker records, per partition on each consumer instance (the consumer
// is any comparable value - normally the *kafka.Consumer pointer itself), when it was
// last revoked, so a later AssignedPartitions event that hands the same partition back
// to the same consumer can report how long that partition went unowned by it.
//
// Tracking per partition (not per consumer) matters under cooperative rebalancing: a
// member keeps most of its partitions and only gives up a few, so "time since this
// consumer last had a revoke" says nothing about the partitions that never stopped.
//
// This measures the pause at its source, with whatever timestamp resolution Go's
// time.Now() gives, instead of inferring it from periodic `kafka-consumer-groups
// --describe` polling, which can miss short pauses entirely.
//
// Exposed as its own type (rather than a bare package-level map) so tests can
// construct an isolated instance instead of sharing defaultPauseTracker's state
// across test cases.
type rebalancePauseTracker struct {
	mu        sync.Mutex
	revokedAt map[pauseKey]time.Time
}

func newRebalancePauseTracker() *rebalancePauseTracker {
	return &rebalancePauseTracker{revokedAt: make(map[pauseKey]time.Time)}
}

// revoked remembers the revocation time of each partition for key, and drops entries
// older than pauseTrackerTTL so partitions that never come back do not accumulate.
func (t *rebalancePauseTracker) revoked(key any, partitions []kafka.TopicPartition, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, ts := range t.revokedAt {
		if at.Sub(ts) > pauseTrackerTTL {
			delete(t.revokedAt, k)
		}
	}
	for _, tp := range partitions {
		t.revokedAt[newPauseKey(key, tp)] = at
	}
}

// assigned returns one pause per assigned partition that was previously revoked on the
// same key, and forgets them. Partitions with no matching revoke (e.g. the first
// assignment a consumer ever receives) and entries older than pauseTrackerTTL yield
// nothing.
func (t *rebalancePauseTracker) assigned(key any, partitions []kafka.TopicPartition, at time.Time) []time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	var pauses []time.Duration
	for _, tp := range partitions {
		k := newPauseKey(key, tp)
		revokedAt, had := t.revokedAt[k]
		if !had {
			continue
		}
		delete(t.revokedAt, k)
		if d := at.Sub(revokedAt); d <= pauseTrackerTTL {
			pauses = append(pauses, d)
		}
	}
	return pauses
}

func maxDuration(ds []time.Duration) time.Duration {
	var m time.Duration
	for _, d := range ds {
		if d > m {
			m = d
		}
	}
	return m
}

// defaultPauseTracker is shared by every rebalanceCb closure created via
// createConsumer, which is fine - it's keyed per-consumer, so multiple consumers in
// the same process (ConsumerConfig.ConsumerCount > 1) don't interfere with each
// other's pause measurements.
var defaultPauseTracker = newRebalancePauseTracker()

// Note on member_id: librdkafka's Go binding does not expose the broker-assigned
// group member.id (the "rdkafka-<uuid>" string visible in `kafka-consumer-groups
// --describe` output) through a public *kafka.Consumer method. (*kafka.Consumer).String()
// (the client's own generated instance name, e.g. "rdkafka#consumer-1") is used
// instead as the closest available stable per-process identifier in the log fields
// below - sufficient to correlate one process's own revoke/assign pairs and to tell
// one pod's log lines from another's, but it is NOT the same string as the
// CONSUMER-ID column in --describe output.

// assignedPartitionsLogFields builds the structured log fields for a
// kafka.AssignedPartitions event. Factored out of rebalanceCb so the exact same field
// set/shape is exercised by both the live callback and rebalance_test.go's
// invokeRebalanceCb mirror, without needing a live *kafka.Consumer in tests.
func assignedPartitionsLogFields(now time.Time, memberID string, partitions []kafka.TopicPartition, pause time.Duration, hadPause bool) map[string]interface{} {
	fields := map[string]interface{}{
		"event":               "partitions_assigned",
		"timestamp_ns":        now.UnixNano(),
		"timestamp":           now.UTC().Format(time.RFC3339Nano),
		"member_id":           memberID,
		"assigned_count":      len(partitions),
		"assigned_partitions": partitions,
	}
	if hadPause {
		fields["pause_duration_ns"] = pause.Nanoseconds()
		fields["pause_duration_ms"] = float64(pause.Nanoseconds()) / 1e6
	}
	return fields
}

// revokedPartitionsLogFields is assignedPartitionsLogFields' counterpart for
// kafka.RevokedPartitions events.
func revokedPartitionsLogFields(now time.Time, memberID string, partitions []kafka.TopicPartition) map[string]interface{} {
	return map[string]interface{}{
		"event":              "partitions_revoked",
		"timestamp_ns":       now.UnixNano(),
		"timestamp":          now.UTC().Format(time.RFC3339Nano),
		"member_id":          memberID,
		"revoked_count":      len(partitions),
		"revoked_partitions": partitions,
	}
}

// rebalanceCb returns a kafka.RebalanceCb for use with SubscribeTopics. It supports
// both rebalance protocols: with a cooperative assignor ("cooperative-sticky") it
// applies incremental assign/unassign deltas, with an eager one (roundrobin, range)
// it does a full Assign/Unassign. See handleRebalance.
func rebalanceCb(logger rebalanceLogger, group string) func(*kafka.Consumer, kafka.Event) error {
	return func(c *kafka.Consumer, ev kafka.Event) error {
		return handleRebalance(logger, defaultPauseTracker, group, c, c.String(), c, ev)
	}
}

// rebalanceActor is the subset of *kafka.Consumer that handleRebalance needs.
type rebalanceActor interface {
	Commit() ([]kafka.TopicPartition, error)
	Assign([]kafka.TopicPartition) error
	Unassign() error
	IncrementalAssign([]kafka.TopicPartition) error
	IncrementalUnassign([]kafka.TopicPartition) error
	GetRebalanceProtocol() string
}

// handleRebalance picks the assignment API from the protocol the group actually
// negotiated, not from our own config: IncrementalAssign/IncrementalUnassign are
// rejected under EAGER, and because calling them marks the event as handled the
// client's own fallback assign never runs, leaving the consumer with no partitions.
func handleRebalance(logger rebalanceLogger, tracker *rebalancePauseTracker, group string, key any, memberID string, c rebalanceActor, ev kafka.Event) error {
	cooperative := c.GetRebalanceProtocol() == "COOPERATIVE"
	protocol := "eager"
	if cooperative {
		protocol = "cooperative"
	}
	switch e := ev.(type) {
	case kafka.AssignedPartitions:
		now := time.Now()
		pauses := tracker.assigned(key, e.Partitions, now)
		prometheus.RebalanceCounter.WithLabelValues(group, protocol, "assigned").Inc()
		for _, p := range pauses {
			prometheus.RebalancePauseHistogram.WithLabelValues(group, protocol).Observe(p.Seconds())
		}
		fields := assignedPartitionsLogFields(now, memberID, e.Partitions, maxDuration(pauses), len(pauses) > 0)
		fields["group"] = group
		fields["protocol"] = protocol
		if len(pauses) > 0 {
			fields["regained_count"] = len(pauses)
		}
		logger.Info("rebalance: partitions assigned", fields)
		if cooperative {
			if err := c.IncrementalAssign(e.Partitions); err != nil {
				return fmt.Errorf("kafka: IncrementalAssign failed: %w", err)
			}
		} else if err := c.Assign(e.Partitions); err != nil {
			return fmt.Errorf("kafka: Assign failed: %w", err)
		}

	case kafka.RevokedPartitions:
		now := time.Now()
		tracker.revoked(key, e.Partitions, now)
		prometheus.RebalanceCounter.WithLabelValues(group, protocol, "revoked").Inc()
		fields := revokedPartitionsLogFields(now, memberID, e.Partitions)
		fields["group"] = group
		fields["protocol"] = protocol
		logger.Info("rebalance: partitions revoked — committing stored offsets", fields)
		// Synchronously flush in-memory stored offsets before handing partitions
		// back: the auto-commit timer may not have fired yet, and a lost stored
		// offset makes the next owner re-process from the last committed position.
		if _, err := c.Commit(); err != nil {
			// ErrNoOffset is benign: nothing was stored on these partitions yet.
			if kErr, ok := err.(kafka.Error); ok && kErr.Code() == kafka.ErrNoOffset {
				prometheus.RevokeCommitErrorsCounter.WithLabelValues(group, "no_offset").Inc()
				logger.Info("rebalance: no stored offsets to commit on revocation (ErrNoOffset — benign)")
			} else {
				// A commit failure must not block the rebalance or the group stalls.
				prometheus.RevokeCommitErrorsCounter.WithLabelValues(group, "error").Inc()
				logger.Error("rebalance: synchronous commit on revocation failed", err)
			}
		}
		if cooperative {
			if len(e.Partitions) > 0 {
				if err := c.IncrementalUnassign(e.Partitions); err != nil {
					return fmt.Errorf("kafka: IncrementalUnassign failed: %w", err)
				}
			}
		} else if err := c.Unassign(); err != nil {
			return fmt.Errorf("kafka: Unassign failed: %w", err)
		}
	}
	return nil
}

// createConsumer builds a *kafka.Consumer, subscribes it to topics with the
// cooperative rebalance callback wired in, and returns it as a confluentConsumer.
//
// The rebalance callback (see rebalanceCb) is what makes cooperative-sticky
// actually work at runtime. Without a non-nil callback, librdkafka silently
// falls back to eager (full-revoke) semantics even when
// partition.assignment.strategy is set to "cooperative-sticky".
func createConsumer(consumerConfig *kafka.ConfigMap, topics []string, log rebalanceLogger) confluentConsumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		panic("error creating consumer:" + err.Error())
	}
	group, _ := (*consumerConfig)["group.id"].(string)
	subscribeErr := consumer.SubscribeTopics(topics, rebalanceCb(log, group))
	if subscribeErr != nil {
		panic("error subscribing to topics:" + subscribeErr.Error())
	}
	return consumer
}

// storeOffsets ensures at least once delivery
// offsets are stored in memory and are later flushed by the auto-commit timer
func storeOffsets(consumer confluentConsumer, partition kafka.TopicPartition) error {
	if partition.Error != nil {
		return fmt.Errorf("error storing offsets:%w", partition.Error)
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}
