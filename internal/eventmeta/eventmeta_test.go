package eventmeta

import (
	"encoding/json"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
)

func TestKafka(t *testing.T) {
	e := &ziggurat.Event{
		RoutingPath: "team/grp/orders/3",
		Metadata:    map[string]any{ziggurat.MetadataKeyKafkaTopic: "orders", ziggurat.MetadataKeyKafkaPartition: 3},
	}
	g, topic, p, ok := Kafka(e)
	if !ok || g != "team/grp" || topic != "orders" || p != 3 {
		t.Fatalf("got (%q,%q,%d,%v)", g, topic, p, ok)
	}
}

// A retried event is JSON round-tripped through RabbitMQ, which turns the partition
// int into a float64. It must still resolve to the same topic and group.
func TestKafka_SurvivesJSONRoundTrip(t *testing.T) {
	orig := ziggurat.Event{
		RoutingPath: "grp/orders/3",
		Metadata:    map[string]any{ziggurat.MetadataKeyKafkaTopic: "orders", ziggurat.MetadataKeyKafkaPartition: 3},
	}
	b, _ := json.Marshal(orig)
	var back ziggurat.Event
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if _, ok := back.Metadata[ziggurat.MetadataKeyKafkaPartition].(float64); !ok {
		t.Fatalf("test premise broken: expected float64 after round trip")
	}
	g, topic, p, ok := Kafka(&back)
	if !ok || g != "grp" || topic != "orders" || p != 3 {
		t.Fatalf("got (%q,%q,%d,%v)", g, topic, p, ok)
	}
}

func TestKafka_NotKafka(t *testing.T) {
	for name, e := range map[string]*ziggurat.Event{
		"no metadata":     {RoutingPath: "a/b/1"},
		"path mismatch":   {RoutingPath: "grp/other/1", Metadata: map[string]any{ziggurat.MetadataKeyKafkaTopic: "orders", ziggurat.MetadataKeyKafkaPartition: 1}},
		"fractional part": {RoutingPath: "g/orders/1", Metadata: map[string]any{ziggurat.MetadataKeyKafkaTopic: "orders", ziggurat.MetadataKeyKafkaPartition: 1.5}},
	} {
		if _, _, _, ok := Kafka(e); ok {
			t.Errorf("%s: expected ok=false", name)
		}
	}
}

func TestRabbitMQQueue(t *testing.T) {
	if _, ok := RabbitMQQueue(&ziggurat.Event{}); ok {
		t.Error("expected no queue on an empty event")
	}
	q, ok := RabbitMQQueue(&ziggurat.Event{Metadata: map[string]any{ziggurat.MetadataKeyRabbitMQQueue: "q"}})
	if !ok || q != "q" {
		t.Errorf("got (%q,%v)", q, ok)
	}
}
