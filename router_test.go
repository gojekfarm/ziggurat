package ziggurat

import (
	"context"
	"reflect"
	"testing"
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
		paths []string
		input []routerEntry
	}{
		{
			name:  "should match the longest prefix without regex",
			want:  "foo.id",
			paths: []string{"foo.id/foo_consumer"},
			input: []routerEntry{{pattern: "foo.id", handler: nil}},
		},
		{
			name:  "should match when topic regex is provided",
			want:  "foo.id/foo_consumer/.*-log",
			paths: []string{"foo.id/foo_consumer/message-log/0", "foo.id/foo_consumer/app-log/0"},
			input: []routerEntry{{pattern: "foo.id/foo_consumer/.*-log", handler: nil}},
		},
		{
			name:  "should match exact partition when a certain partition is specified",
			want:  "",
			paths: []string{"foo.id/foo_consumer/message-log/10"},
			input: []routerEntry{{pattern: "foo.id/foo_consumer/message-log/1$", handler: nil}},
		},
		{
			name:  "should match the longest prefix with regex",
			want:  "foo.id/foo_consumer/.*-log/\\d+",
			paths: []string{"foo.id/foo_consumer/message-log/5"},
			input: []routerEntry{
				{pattern: "foo.id/foo_consumer/.*-log/\\d+", handler: nil},
				{pattern: "foo.id/foo_consumer/.*-log", handler: nil},
			},
		},
		{
			name:  "should not match similar consumer group names",
			want:  "",
			paths: []string{"foo.id/foo_consumer/"},
			input: []routerEntry{
				{pattern: "foo.id/foo/"},
			},
		},
		{
			name:  "should not match similar consumer group names",
			want:  "bar.id/.*",
			paths: []string{"bar.id/GO_BIRD_COMBO-booking-log/11"},
			input: []routerEntry{
				{pattern: "bar.id/.*"},
			},
		},
		{
			name: "regex test: should match paths ending with even numbers",
			want: "bar.id/message-log/(2|4|6|8|10)$",
			input: []routerEntry{
				{pattern: "bar.id/message-log/(1|3|5|7|11)$"},
				{pattern: "bar.id/message-log/(2|4|6|8|10)$"},
			},
			paths: []string{"bar.id/message-log/10"},
		},
	}

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
			for _, es := range test.input {
				r.register(es.pattern, es.handler)
			}
			r.handlerEntry = esToMap(test.input)
			for _, path := range test.paths {
				t.Logf("matching path:%s\n", path)
				_, m := r.match(path)
				t.Logf("match:%s\n", m)
				if m != test.want {
					t.Errorf("%s test failed, expected %q got %q", test.name, test.want, m)
					return
				}
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
	var h HandlerFunc = func(ctx context.Context, event *Event) {

	}
	cases := map[string]func(t *testing.T){
		"should panic on empty pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandlerFunc("", h)
			})
		},
		"should panic on nil handler": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandlerFunc("/bar", nil)
			})
		},

		"should panic on / as the pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandlerFunc("/", h)
			})
		},

		"should not allow multiple registrations of the same pattern": func(t *testing.T) {
			shouldPanic(t, func() {
				router.HandlerFunc("/bar", h)
				router.HandlerFunc("/bar", h)
			})
		},
		"sort and append should append the patterns in increasing order of len(s)": func(t *testing.T) {
			var router Router
			router.HandlerFunc("/bar/baz", h)
			router.HandlerFunc("/bar", h)
			router.HandlerFunc("/foo/bar/0", h)
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
