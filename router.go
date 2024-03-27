package ziggurat

import (
	"context"
	"fmt"
	"regexp"
	"sort"
)

/*
Example route <bootstrap_server>/<topic>/<partition>
routerEntry {
	handler ziggurat.Handler
	pattern string
}
handlerEntry []string sorted by len of paths
*/

// routerEntry contains the pattern and the path routerEntry
type routerEntry struct {
	handler Handler
	pattern string
	rgx     *regexp.Regexp
}

type Router struct {
	handlerEntry map[string]routerEntry
	es           []routerEntry
}

// match works by matching the shortest prefix that matches the path
// it returns the matched path and the handler associated with it
func (r *Router) match(path string) (Handler, string) {
	if e, ok := r.handlerEntry[path]; ok {
		return e.handler, path
	}
	for _, e := range r.es {
		matched := e.rgx.MatchString(path)
		if matched {
			return e.handler, e.pattern
		}

	}
	return nil, ""
}

func sortAndAppend(s []routerEntry, e routerEntry) []routerEntry {
	n := len(s)
	// get the insert position
	// We are sorting all the patterns by len in descending order
	i := sort.Search(n, func(i int) bool {
		return len(e.pattern) > len(s[i].pattern)
	})
	s = append(s, routerEntry{})
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

func (r *Router) HandlerFunc(pattern string, h func(ctx context.Context, event *Event)) {
	if pattern == "" {
		panic(fmt.Errorf("kafka router:pattern cannot be [%q]", pattern))
	}
	if h == nil {
		panic("kafka router:handler cannot be <nil>")
	}
	r.register(pattern, HandlerFunc(h))
}

func (r *Router) register(pattern string, h Handler) {
	if r.handlerEntry == nil {
		r.handlerEntry = make(map[string]routerEntry)
	}

	//check if pattern is `""` OR "/"
	if (len(pattern) == 1 && pattern[len(pattern)-1] == '/') || pattern == "" {
		panic(pattern + " is not a valid pattern")

	}

	//panic on multiple registrations
	if _, ok := r.handlerEntry[pattern]; ok {
		panic(fmt.Sprintf("kafka router:multiple regirstrations for [%s]", pattern))
	}

	e := routerEntry{handler: h, pattern: pattern, rgx: regexp.MustCompile(pattern)}
	r.handlerEntry[pattern] = e

	r.es = sortAndAppend(r.es, e)
}

func (r *Router) Handle(ctx context.Context, event *Event) {
	path := event.RoutingPath
	h, _ := r.match(path)
	if h != nil {
		h.Handle(ctx, event)
	}
	return
}

func NewRouter() *Router {
	return &Router{}
}

type Middleware func(handler Handler) Handler

var pipe = func(h Handler, fs ...Middleware) Handler {
	if len(fs) < 1 {
		return h
	}
	last := len(fs) - 1
	f := func(ctx context.Context, event *Event) {
		next := h
		for i := last; i >= 0; i-- {
			next = fs[i](next)
		}
		next.Handle(ctx, event)
	}
	return HandlerFunc(f)
}

// Use takes a ziggurat.Handler and wraps it with Middleware
func Use(h Handler, fs ...Middleware) Handler {
	return pipe(h, fs...)
}
