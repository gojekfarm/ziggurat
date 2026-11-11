package kafka

import (
	"errors"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Config tests: partition.assignment.strategy default and override
// ---------------------------------------------------------------------------

func TestToConfigMap_DefaultPartitionAssignment_IsCooperativeSticky(t *testing.T) {
	// When PartitionAssignment is left empty, toConfigMap must set the strategy
	// to DefaultPartitionAssignmentStrategy ("cooperative-sticky") so that the
	// librdkafka client uses incremental rebalancing without any caller having
	// to opt in explicitly.
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
	}.toConfigMap()

	assertConfigValue(t, cm, "partition.assignment.strategy", DefaultPartitionAssignmentStrategy)
}

func TestToConfigMap_ExplicitPartitionAssignment_OverridesDefault(t *testing.T) {
	// A caller that explicitly sets PartitionAssignment must be able to override
	// the default — e.g. to roll back to "range" during a canary or incident.
	const override = "range,roundrobin"
	cm := ConsumerConfig{
		BootstrapServers:    "localhost:9092",
		GroupID:             "g1",
		PartitionAssignment: override,
	}.toConfigMap()

	assertConfigValue(t, cm, "partition.assignment.strategy", override)
}

// ---------------------------------------------------------------------------
// Rebalance callback tests
//
// rebalanceCb returns a func(*kafka.Consumer, kafka.Event) error, but a real
// *kafka.Consumer requires a broker connection — impossible in a unit test.
// We therefore exercise the callback *logic* through the invokeRebalanceCb
// adapter (defined below) which replaces *kafka.Consumer with the narrow
// rebalanceConsumerActions interface that our MockConsumer satisfies.
// ---------------------------------------------------------------------------

// rebalanceLogCapture records Info/Error calls so tests can assert on them.
type rebalanceLogCapture struct {
	infos      []string
	errors     []string
	infoFields []map[string]interface{} // fields passed to the most recent calls, 1:1 with infos
}

func (l *rebalanceLogCapture) Info(msg string, fields ...map[string]interface{}) {
	l.infos = append(l.infos, msg)
	if len(fields) > 0 {
		l.infoFields = append(l.infoFields, fields[0])
	} else {
		l.infoFields = append(l.infoFields, nil)
	}
}

func (l *rebalanceLogCapture) Error(msg string, _ error, _ ...map[string]interface{}) {
	l.errors = append(l.errors, msg)
}

// rebalanceMockConsumer wraps MockConsumer and adds mock implementations for
// the rebalance-specific methods so testify can record/assert on them.
type rebalanceMockConsumer struct {
	mock.Mock
	protocol string
}

func (m *rebalanceMockConsumer) IncrementalAssign(partitions []kafka.TopicPartition) error {
	return m.Called(partitions).Error(0)
}

func (m *rebalanceMockConsumer) IncrementalUnassign(partitions []kafka.TopicPartition) error {
	return m.Called(partitions).Error(0)
}

func (m *rebalanceMockConsumer) Commit() ([]kafka.TopicPartition, error) {
	args := m.Called()
	return args.Get(0).([]kafka.TopicPartition), args.Error(1)
}

// rebalanceMockConsumer implements rebalanceActor. protocol is what
// GetRebalanceProtocol reports; empty means the cooperative-sticky default these
// tests were written for.
func (m *rebalanceMockConsumer) Assign(partitions []kafka.TopicPartition) error {
	return m.Called(partitions).Error(0)
}

func (m *rebalanceMockConsumer) Unassign() error {
	return m.Called().Error(0)
}

func (m *rebalanceMockConsumer) GetRebalanceProtocol() string {
	if m.protocol == "" {
		return "COOPERATIVE"
	}
	return m.protocol
}

// invokeRebalanceCb runs the production handleRebalance against a mock consumer.
// tracker and key let a test isolate pause-tracking state.
func invokeRebalanceCb(log rebalanceLogger, tracker *rebalancePauseTracker, key any, memberID string, c rebalanceActor, ev kafka.Event) error {
	return handleRebalance(log, tracker, key, memberID, c, ev)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRebalanceCb_AssignedPartitions_CallsIncrementalAssign(t *testing.T) {
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	partitions := []kafka.TopicPartition{
		{Topic: makePtr("foo"), Partition: 0},
		{Topic: makePtr("foo"), Partition: 1},
	}
	mc.On("IncrementalAssign", partitions).Return(nil)

	ev := kafka.AssignedPartitions{Partitions: partitions}
	if err := invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev); err != nil {
		t.Fatalf("expected no error from AssignedPartitions callback, got: %v", err)
	}
	mc.AssertCalled(t, "IncrementalAssign", partitions)
	mc.AssertNotCalled(t, "Commit")
	mc.AssertNotCalled(t, "IncrementalUnassign", mock.Anything)
}

func TestRebalanceCb_RevokedPartitions_CommitsThenUnassigns(t *testing.T) {
	// Commit must be called before IncrementalUnassign: flushing stored offsets
	// synchronously before the partitions leave this consumer prevents the next
	// owner from re-processing already-handled messages.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	partitions := []kafka.TopicPartition{
		{Topic: makePtr("bar"), Partition: 2},
	}
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("IncrementalUnassign", partitions).Return(nil)

	ev := kafka.RevokedPartitions{Partitions: partitions}
	if err := invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev); err != nil {
		t.Fatalf("expected no error from RevokedPartitions callback, got: %v", err)
	}

	mc.AssertCalled(t, "Commit")
	mc.AssertCalled(t, "IncrementalUnassign", partitions)
}

func TestRebalanceCb_RevokedPartitions_EmptySlice_SkipsUnassign(t *testing.T) {
	// During a rolling upgrade from eager to cooperative, librdkafka delivers a
	// RevokedPartitions event with an empty Partitions slice to signal "revoke
	// everything". The callback must still commit but must NOT call
	// IncrementalUnassign with an empty slice.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)

	ev := kafka.RevokedPartitions{Partitions: []kafka.TopicPartition{}}
	if err := invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	mc.AssertCalled(t, "Commit")
	mc.AssertNotCalled(t, "IncrementalUnassign", mock.Anything)
}

func TestRebalanceCb_RevokedPartitions_ErrNoOffset_IsNotLoggedAsError(t *testing.T) {
	// ErrNoOffset from Commit means there were no stored offsets to flush —
	// this is benign and must be logged at Info level, not Error.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	noOffsetErr := kafka.NewError(kafka.ErrNoOffset, "no offset to commit", false)
	mc.On("Commit").Return([]kafka.TopicPartition{}, noOffsetErr)
	mc.On("IncrementalUnassign", mock.Anything).Return(nil)

	partitions := []kafka.TopicPartition{{Topic: makePtr("baz"), Partition: 0}}
	ev := kafka.RevokedPartitions{Partitions: partitions}
	if err := invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev); err != nil {
		t.Fatalf("expected no error for ErrNoOffset, got: %v", err)
	}

	if len(log.errors) > 0 {
		t.Errorf("ErrNoOffset must not be logged at Error level; got error logs: %v", log.errors)
	}
}

func TestRebalanceCb_RevokedPartitions_CommitError_IsLoggedButRebalanceContinues(t *testing.T) {
	// A non-ErrNoOffset commit failure must be logged as an error but the
	// callback must still call IncrementalUnassign — a commit failure must not
	// stall the rebalance.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	mc.On("Commit").Return([]kafka.TopicPartition{}, errors.New("broker unavailable"))
	mc.On("IncrementalUnassign", mock.Anything).Return(nil)

	partitions := []kafka.TopicPartition{{Topic: makePtr("qux"), Partition: 1}}
	ev := kafka.RevokedPartitions{Partitions: partitions}
	_ = invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev)

	if len(log.errors) == 0 {
		t.Error("expected a non-ErrNoOffset commit failure to be logged at Error level")
	}
	mc.AssertCalled(t, "IncrementalUnassign", partitions)
}

// ---------------------------------------------------------------------------
// Pause-duration / structured-field tests (t401572: rebalance listener logging)
// ---------------------------------------------------------------------------

func TestRebalanceCb_AssignedPartitions_FirstEverAssignment_HasNoPauseDuration(t *testing.T) {
	// A consumer's very first AssignedPartitions (no prior RevokedPartitions on
	// this key) must not report a pause_duration - there is nothing to measure it
	// from, and reporting a bogus 0 would be indistinguishable from "resolved
	// instantly", which is a real and different outcome.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	partitions := []kafka.TopicPartition{{Topic: makePtr("foo"), Partition: 0}}
	mc.On("IncrementalAssign", partitions).Return(nil)

	ev := kafka.AssignedPartitions{Partitions: partitions}
	if err := invokeRebalanceCb(log, newRebalancePauseTracker(), "k1", "member-1", mc, ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := log.infoFields[len(log.infoFields)-1]
	if _, present := fields["pause_duration_ms"]; present {
		t.Errorf("expected no pause_duration_ms on a first-ever assignment, got fields: %+v", fields)
	}
	if fields["event"] != "partitions_assigned" {
		t.Errorf(`expected fields["event"] == "partitions_assigned", got %v`, fields["event"])
	}
	if fields["member_id"] != "member-1" {
		t.Errorf(`expected fields["member_id"] == "member-1", got %v`, fields["member_id"])
	}
}

func TestRebalanceCb_AssignedPartitions_AfterRevoke_ReportsPauseDuration(t *testing.T) {
	// The core requirement this test guards: a RevokedPartitions event followed by
	// an AssignedPartitions event on the SAME key must report a pause_duration_ms
	// computed from the wall-clock gap between them.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	tracker := newRebalancePauseTracker()

	revoked := []kafka.TopicPartition{{Topic: makePtr("foo"), Partition: 0}}
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("IncrementalUnassign", revoked).Return(nil)
	if err := invokeRebalanceCb(log, tracker, "k1", "member-1", mc, kafka.RevokedPartitions{Partitions: revoked}); err != nil {
		t.Fatalf("unexpected error on revoke: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	assigned := []kafka.TopicPartition{{Topic: makePtr("foo"), Partition: 0}}
	mc.On("IncrementalAssign", assigned).Return(nil)
	if err := invokeRebalanceCb(log, tracker, "k1", "member-1", mc, kafka.AssignedPartitions{Partitions: assigned}); err != nil {
		t.Fatalf("unexpected error on assign: %v", err)
	}

	fields := log.infoFields[len(log.infoFields)-1]
	pauseMs, ok := fields["pause_duration_ms"].(float64)
	if !ok {
		t.Fatalf("expected fields[\"pause_duration_ms\"] to be a float64, got %#v", fields["pause_duration_ms"])
	}
	if pauseMs < 5.0 {
		t.Errorf("expected pause_duration_ms >= 5 (slept 5ms between revoke and assign), got %v", pauseMs)
	}
}

func TestRebalanceCb_PauseTracker_IsolatedPerKey(t *testing.T) {
	// Two different consumer keys (two members in the same process, e.g.
	// ConsumerConfig.ConsumerCount > 1) must not leak pause state into each other:
	// a revoke on one key must not produce a pause_duration on an assign for a
	// different key.
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{}
	tracker := newRebalancePauseTracker()

	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("IncrementalUnassign", mock.Anything).Return(nil)
	mc.On("IncrementalAssign", mock.Anything).Return(nil)

	if err := invokeRebalanceCb(log, tracker, "consumer-a", "member-a", mc, kafka.RevokedPartitions{Partitions: nil}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := invokeRebalanceCb(log, tracker, "consumer-b", "member-b", mc, kafka.AssignedPartitions{Partitions: nil}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := log.infoFields[len(log.infoFields)-1]
	if _, present := fields["pause_duration_ms"]; present {
		t.Errorf("expected no pause_duration_ms for consumer-b (only consumer-a was revoked), got fields: %+v", fields)
	}
}

func TestRebalanceCb_Eager_AssignedPartitions_CallsAssignNotIncremental(t *testing.T) {
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{protocol: "EAGER"}
	partitions := []kafka.TopicPartition{{Partition: 0}, {Partition: 1}}
	mc.On("Assign", partitions).Return(nil)

	err := invokeRebalanceCb(log, newRebalancePauseTracker(), "c1", "m1", mc, kafka.AssignedPartitions{Partitions: partitions})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mc.AssertCalled(t, "Assign", partitions)
	mc.AssertNotCalled(t, "IncrementalAssign", partitions)
}

func TestRebalanceCb_Eager_RevokedPartitions_CommitsThenUnassignsAll(t *testing.T) {
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{protocol: "EAGER"}
	partitions := []kafka.TopicPartition{{Partition: 0}, {Partition: 1}}
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("Unassign").Return(nil)

	err := invokeRebalanceCb(log, newRebalancePauseTracker(), "c1", "m1", mc, kafka.RevokedPartitions{Partitions: partitions})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mc.AssertCalled(t, "Commit")
	mc.AssertCalled(t, "Unassign")
	mc.AssertNotCalled(t, "IncrementalUnassign", partitions)
}

func TestRebalanceCb_Eager_PauseDurationIsReported(t *testing.T) {
	log := &rebalanceLogCapture{}
	mc := &rebalanceMockConsumer{protocol: "EAGER"}
	tracker := newRebalancePauseTracker()
	mc.On("Commit").Return([]kafka.TopicPartition{}, nil)
	mc.On("Unassign").Return(nil)
	mc.On("Assign", mock.Anything).Return(nil)

	_ = invokeRebalanceCb(log, tracker, "c1", "m1", mc, kafka.RevokedPartitions{Partitions: []kafka.TopicPartition{{Partition: 0}}})
	_ = invokeRebalanceCb(log, tracker, "c1", "m1", mc, kafka.AssignedPartitions{Partitions: []kafka.TopicPartition{{Partition: 0}}})

	last := log.infoFields[len(log.infoFields)-1]
	if _, ok := last["pause_duration_ms"]; !ok {
		t.Fatalf("expected pause_duration_ms on assignment after a revoke, got %v", last)
	}
}
