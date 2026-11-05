package kafka

import (
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// rebalanceLogger is the subset of ziggurat.StructuredLogger that the rebalance
// callback needs. Kept as a local interface so the kafka package does not need to
// import the top-level ziggurat package just for its logger interface, avoiding a
// circular dependency.
type rebalanceLogger interface {
	Error(string, error, ...map[string]interface{})
	Info(string, ...map[string]interface{})
}

// rebalancePauseTracker records, per consumer instance (keyed by any comparable
// value - normally the *kafka.Consumer pointer itself), the timestamp of the most
// recent RevokedPartitions event, so a later AssignedPartitions event on the same
// key can report how long that member went without actively holding any partitions.
//
// This measures the pause at its source, with whatever timestamp resolution Go's
// time.Now() gives (effectively sub-microsecond on any platform this runs on),
// instead of inferring it from periodic `kafka-consumer-groups --describe` polling,
// which is typically only run every several seconds and can miss short pauses
// entirely or report their duration with only poll-interval precision.
//
// Exposed as its own type (rather than a bare package-level map) so tests can
// construct an isolated instance instead of sharing defaultPauseTracker's state
// across test cases.
type rebalancePauseTracker struct {
	mu        sync.Mutex
	revokedAt map[any]time.Time
}

func newRebalancePauseTracker() *rebalancePauseTracker {
	return &rebalancePauseTracker{revokedAt: make(map[any]time.Time)}
}

func (t *rebalancePauseTracker) revoked(key any, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.revokedAt[key] = at
}

// assigned reports the elapsed time since the most recent revoked() call for key,
// and clears it. ok is false if there was no matching revoked() call (e.g. the very
// first assignment this consumer ever receives, with no prior revocation).
func (t *rebalancePauseTracker) assigned(key any, at time.Time) (pause time.Duration, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	revokedAt, had := t.revokedAt[key]
	if !had {
		return 0, false
	}
	delete(t.revokedAt, key)
	return at.Sub(revokedAt), true
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
func rebalanceCb(logger rebalanceLogger) func(*kafka.Consumer, kafka.Event) error {
	return func(c *kafka.Consumer, ev kafka.Event) error {
		return handleRebalance(logger, defaultPauseTracker, c, c.String(), c, ev)
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
func handleRebalance(logger rebalanceLogger, tracker *rebalancePauseTracker, key any, memberID string, c rebalanceActor, ev kafka.Event) error {
	cooperative := c.GetRebalanceProtocol() == "COOPERATIVE"
	switch e := ev.(type) {
	case kafka.AssignedPartitions:
		now := time.Now()
		pause, hadPause := tracker.assigned(key, now)
		logger.Info("rebalance: partitions assigned", assignedPartitionsLogFields(now, memberID, e.Partitions, pause, hadPause))
		if cooperative {
			if err := c.IncrementalAssign(e.Partitions); err != nil {
				return fmt.Errorf("kafka: IncrementalAssign failed: %w", err)
			}
		} else if err := c.Assign(e.Partitions); err != nil {
			return fmt.Errorf("kafka: Assign failed: %w", err)
		}

	case kafka.RevokedPartitions:
		now := time.Now()
		tracker.revoked(key, now)
		logger.Info("rebalance: partitions revoked — committing stored offsets", revokedPartitionsLogFields(now, memberID, e.Partitions))
		// Synchronously flush in-memory stored offsets before handing partitions
		// back: the auto-commit timer may not have fired yet, and a lost stored
		// offset makes the next owner re-process from the last committed position.
		if _, err := c.Commit(); err != nil {
			// ErrNoOffset is benign: nothing was stored on these partitions yet.
			if kErr, ok := err.(kafka.Error); ok && kErr.Code() == kafka.ErrNoOffset {
				logger.Info("rebalance: no stored offsets to commit on revocation (ErrNoOffset — benign)")
			} else {
				// A commit failure must not block the rebalance or the group stalls.
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
	subscribeErr := consumer.SubscribeTopics(topics, rebalanceCb(log))
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
