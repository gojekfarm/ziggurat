package kafka

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat"
)

func Test_match(t *testing.T) {
	kr := NewRouter()
	f := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error { return nil })
	sortedPaths := []string{"localhost:9092", "localhost:9092/foo_consumer", "localhost:9092/foo_consumer/bar_topic"}
	var es []entry
	entries := make(map[string]entry, len(sortedPaths))
	for _, sp := range sortedPaths {
		e := entry{
			handler: f,
			pattern: sp,
		}
		entries[sp] = e
		es = append(es, e)
	}
	kr.es = es
	kr.entries = entries

	_, m := kr.match("localhost:9092/foo_consumer/bar_topic/0")
	if m != sortedPaths[0] {
		t.Errorf("expected match to be %s but got %s", es[0], m)
	}
}
