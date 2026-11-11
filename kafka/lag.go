package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"
)

// DefaultStatisticsIntervalMS is how often librdkafka emits its statistics event, which
// ziggurat turns into the ziggurat_go_kafka_consumer_lag gauge. See
// ConsumerConfig.StatisticsIntervalMS.
const DefaultStatisticsIntervalMS = 15000

// statsPayload is the part of librdkafka's statistics JSON (see
// https://github.com/confluentinc/librdkafka/blob/master/STATISTICS.md) used for lag.
type statsPayload struct {
	Topics map[string]struct {
		Partitions map[string]struct {
			Partition   int32 `json:"partition"`
			ConsumerLag int64 `json:"consumer_lag"`
		} `json:"partitions"`
	} `json:"topics"`
}

type lagKey struct {
	topic     string
	partition int32
}

// lagReporter turns the statistics events of ONE consumer into consumer_lag gauge
// values for the partitions that consumer currently owns, and removes the series of
// partitions it no longer owns. Without that removal a rebalance would leave stale
// series behind on the old owner, reporting a frozen lag forever. It is used by a
// single worker goroutine, so it needs no locking.
type lagReporter struct {
	group string
	prev  map[lagKey]struct{}
}

func newLagReporter(group string) *lagReporter {
	return &lagReporter{group: group, prev: make(map[lagKey]struct{})}
}

// update parses one statistics JSON document and refreshes the gauges. librdkafka
// reports consumer_lag -1 for partitions it is not fetching (not owned, or lag not yet
// known) and lists the internal partition -1 for each topic; both are skipped.
func (r *lagReporter) update(statsJSON string) error {
	if statsJSON == "" {
		return errors.New("empty statistics payload")
	}
	var p statsPayload
	if err := json.Unmarshal([]byte(statsJSON), &p); err != nil {
		return fmt.Errorf("kafka: decoding statistics: %w", err)
	}

	current := make(map[lagKey]struct{})
	for topic, t := range p.Topics {
		for _, part := range t.Partitions {
			if part.Partition < 0 || part.ConsumerLag < 0 {
				continue
			}
			k := lagKey{topic: topic, partition: part.Partition}
			current[k] = struct{}{}
			prometheus.ConsumerLagGauge.WithLabelValues(r.group, topic, strconv.Itoa(int(part.Partition))).Set(float64(part.ConsumerLag))
		}
	}
	for k := range r.prev {
		if _, still := current[k]; !still {
			prometheus.ConsumerLagGauge.DeleteLabelValues(r.group, k.topic, strconv.Itoa(int(k.partition)))
		}
	}
	r.prev = current
	return nil
}

// clear removes every series this reporter created. Called when the worker stops.
func (r *lagReporter) clear() {
	for k := range r.prev {
		prometheus.ConsumerLagGauge.DeleteLabelValues(r.group, k.topic, strconv.Itoa(int(k.partition)))
	}
	r.prev = make(map[lagKey]struct{})
}
