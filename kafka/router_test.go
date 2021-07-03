package kafka

import (
	"context"
	"reflect"
	"testing"

	"github.com/gojekfarm/ziggurat"
)

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func Test_match(t *testing.T) {
	tests := []struct {
		want  string
		name  string
		path  string
		input []routerEntry
	}{
		{
			name:  "should match the longest prefix without regex",
			want:  "localhost:9092",
			path:  "localhost:9092/foo_consumer",
			input: []routerEntry{{pattern: "localhost:9092", handler: nil}},
		},
		{
			name:  "should match when topic regex is provided",
			want:  "localhost:9092/foo_consumer/.*-log",
			path:  "localhost:9092/foo_consumer/message-log/0",
			input: []routerEntry{{pattern: "localhost:9092/foo_consumer/.*-log", handler: nil}},
		},
		{
			name:  "should match exact partition when a certain partition is specified",
			want:  "",
			path:  "localhost:9092/foo_consumer/message-log/10",
			input: []routerEntry{{pattern: "localhost:9092/foo_consumer/message-log/1$", handler: nil}},
		},
		{
			name: "should match the longest prefix with regex",
			want: "localhost:9092/foo_consumer/.*-log/\\d+",
			path: "localhost:9092/foo_consumer/message-log/5",
			input: []routerEntry{
				{pattern: "localhost:9092/foo_consumer/.*-log/\\d+", handler: nil},
				{pattern: "localhost:9092/foo_consumer/.*-log", handler: nil},
			},
		}}

	esToMap := func(es []routerEntry) map[string]routerEntry {
		m := make(map[string]routerEntry, len(es))
		for _, e := range es {
			m[e.pattern] = routerEntry{pattern: e.pattern}
		}
		return m
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var r Router
			r.es = test.input
			r.handlerEntry = esToMap(test.input)
			_, m := r.match(test.path)
			if m != test.want {
				t.Errorf("%s test failed, expected %q got %q", test.name, test.want, m)
			}
		})
	}
}

func Test_sortAndAppend(t *testing.T) {
	var es []routerEntry
	want := []routerEntry{
		{pattern: "foo/bar/baz/2"},
		{pattern: "foo/bar/0"},
		{pattern: "bar/baz"},
	}

	patterns := []string{"foo/bar/0", "bar/baz", "foo/bar/baz/2"}

	for _, p := range patterns {
		es = sortAndAppend(es, routerEntry{pattern: p})
	}

	if !reflect.DeepEqual(want, es) {
		t.Errorf("expected %v but got %v", want, es)
	}
}

func Test_register(t *testing.T) {
	var router Router
	var h ziggurat.HandlerFunc = func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	}
	cases := map[string]func(t *testing.T){
		"should panic on empty pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandleFunc("", h)
			})
		},
		"should panic on nil handler": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandleFunc("/bar", nil)
			})
		},

		"should panic on / as the pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandleFunc("/", h)
			})
		},

		"should not allow multiple registrations of the same pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandleFunc("/bar", h)
				router.HandleFunc("/bar", h)
			})
		},
		"sort and append should append the patterns in increasing order of len(s)": func(t *testing.T) {
			var router Router
			router.HandleFunc("/bar/baz", h)
			router.HandleFunc("/bar", h)
			router.HandleFunc("/foo/bar/0", h)
			want := []routerEntry{
				{pattern: "/foo/bar/0", handler: h},
				{pattern: "/bar/baz", handler: h},
				{pattern: "/bar", handler: h},
			}
			if !reflect.DeepEqual(want, router.es) {
				for i, w := range want {
					if w.pattern != router.es[i].pattern {
						t.Errorf("expected: %s, got %s", w.pattern, router.es[i].pattern)
					}
				}
			}
		},
	}
	for c, f := range cases {
		t.Run(c, f)
	}
}
