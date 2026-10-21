package kafka

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2"
)

func TestHeadersToMap(t *testing.T) {
	t.Run("nil for no headers", func(t *testing.T) {
		if got := headersToMap(nil); got != nil {
			t.Errorf("expected nil for no headers, got %v", got)
		}
		if got := headersToMap([]kafka.Header{}); got != nil {
			t.Errorf("expected nil for empty headers, got %v", got)
		}
	})

	t.Run("converts headers, last value wins on duplicate key", func(t *testing.T) {
		headers := []kafka.Header{
			{Key: "traceparent", Value: []byte("00-abc-def-01")},
			{Key: "dup", Value: []byte("first")},
			{Key: "dup", Value: []byte("second")},
		}
		got := headersToMap(headers)
		want := map[string]string{
			"traceparent": "00-abc-def-01",
			"dup":         "second",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
	})
}

func TestProcessMessage_CapturesHeadersIntoMetadata(t *testing.T) {
	var gotEvent *ziggurat.Event
	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		gotEvent = e
	})

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: makePtr("foo"), Partition: 1, Offset: 1},
		Key:            []byte("k1"),
		Value:          []byte("v1"),
		Timestamp:      time.Now(),
		Headers: []kafka.Header{
			{Key: "traceparent", Value: []byte("00-abc-def-01")},
		},
	}

	processMessage(context.Background(), msg, handler, "test-group")

	if gotEvent == nil {
		t.Fatal("expected handler to receive an event")
	}
	headers, ok := gotEvent.Metadata[ziggurat.MetadataKeyHeaders].(map[string]string)
	if !ok {
		t.Fatalf("expected Metadata[%q] to be a map[string]string, got %T", ziggurat.MetadataKeyHeaders, gotEvent.Metadata[ziggurat.MetadataKeyHeaders])
	}
	if headers["traceparent"] != "00-abc-def-01" {
		t.Errorf("expected traceparent header to be captured, got %v", headers)
	}
}

func TestProcessMessage_NoHeaders(t *testing.T) {
	var gotEvent *ziggurat.Event
	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		gotEvent = e
	})

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: makePtr("foo"), Partition: 1, Offset: 1},
		Key:            []byte("k1"),
		Value:          []byte("v1"),
		Timestamp:      time.Now(),
	}

	processMessage(context.Background(), msg, handler, "test-group")

	if gotEvent == nil {
		t.Fatal("expected handler to receive an event")
	}
	if _, ok := gotEvent.Metadata[ziggurat.MetadataKeyHeaders]; ok {
		t.Errorf("expected no headers key at all when message has none, got %v", gotEvent.Metadata[ziggurat.MetadataKeyHeaders])
	}
}
