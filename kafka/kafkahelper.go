package kafka

import (
	"fmt"

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

// rebalanceCb returns a kafka.RebalanceCb for use with SubscribeTopics.
//
// When the consumer group uses a cooperative assignor ("cooperative-sticky"),
// the broker sends incremental assignment/revocation events rather than
// revoking ALL partitions and re-assigning from scratch on every membership
// change. This callback handles both sides of that protocol:
//
//   - AssignedPartitions: calls IncrementalAssign so that only newly-granted
//     partitions are added to this consumer's assignment without touching the
//     ones it is already consuming. librdkafka will not start fetching from
//     the new partitions until IncrementalAssign is called.
//
//   - RevokedPartitions: synchronously commits any in-memory stored offsets
//     before calling IncrementalUnassign. The synchronous commit is the key
//     safety step: the auto-commit timer (auto.commit.interval.ms) may not have
//     fired yet, so StoreOffsets-only offsets would be lost if we handed the
//     partitions back to the coordinator without flushing them first. A lost
//     stored offset means the next owner of that partition restarts from the
//     last *committed* position, producing duplicates.
//
// If the group is running with an eager assignor (e.g. during a rolling upgrade
// where some members have not yet switched to cooperative-sticky), librdkafka
// delivers a RevokedPartitions event with an empty Partitions slice to signal
// "revoke everything". The nil-check on ev.Partitions handles that case: we
// still commit synchronously but skip the no-op IncrementalUnassign call.
func rebalanceCb(logger rebalanceLogger) func(*kafka.Consumer, kafka.Event) error {
	return func(c *kafka.Consumer, ev kafka.Event) error {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			logger.Info("rebalance: partitions assigned", map[string]interface{}{
				"count":      len(e.Partitions),
				"partitions": e.Partitions,
			})
			if err := c.IncrementalAssign(e.Partitions); err != nil {
				return fmt.Errorf("kafka: IncrementalAssign failed: %w", err)
			}

		case kafka.RevokedPartitions:
			logger.Info("rebalance: partitions revoked — committing stored offsets", map[string]interface{}{
				"count":      len(e.Partitions),
				"partitions": e.Partitions,
			})
			// Synchronously flush in-memory stored offsets before handing
			// partitions back to the coordinator. This closes the window where
			// the auto-commit timer has not yet fired but partitions are being
			// moved away, which would otherwise produce duplicates for the next
			// consumer that picks them up.
			if _, err := c.Commit(); err != nil {
				// ErrNoOffset is benign: it means there were no stored offsets
				// to commit (e.g. no messages have been processed yet on these
				// partitions). Treat it as a non-fatal informational event.
				if kErr, ok := err.(kafka.Error); ok && kErr.Code() == kafka.ErrNoOffset {
					logger.Info("rebalance: no stored offsets to commit on revocation (ErrNoOffset — benign)")
				} else {
					logger.Error("rebalance: synchronous commit on revocation failed", err)
					// Return the error so librdkafka can log it; the unassign
					// still proceeds — a commit failure must not block the
					// rebalance or the group will stall.
				}
			}
			if len(e.Partitions) > 0 {
				if err := c.IncrementalUnassign(e.Partitions); err != nil {
					return fmt.Errorf("kafka: IncrementalUnassign failed: %w", err)
				}
			}
		}
		return nil
	}
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
