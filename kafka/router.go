package kafka

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/gojekfarm/ziggurat"
)

/*
Example route <bootstrap_server>/<topic>/<partition>
routerEntry {
	handler ziggurat.Handler
	pattern string
}
handlerEntry []string sorted by len of paths
*/

//routerEntry contains the pattern and the path routerEntry
type routerEntry struct {
	handler ziggurat.Handler
	pattern string
}

type Router struct {
	handlerEntry map[string]routerEntry
	es           []routerEntry
}

//match works by matching the shortest prefix that matches the path
// it returns the matched path and the handler associated with it
func (r *Router) match(path string) (ziggurat.Handler, string) {

	if e, ok := r.handlerEntry[path]; ok {
		return e.handler, path
	}
	for _, e := range r.es {
		//		if strings.HasPrefix(path, e.pattern) {
		//			return e.handler, e.pattern
		//		}
		matched, err := regexp.MatchString(e.pattern, path)
		if err != nil {
			panic(err)
		}

		if matched {
			return e.handler, e.pattern
		}

	}

	return nil, ""
}

func sortAndAppend(s []routerEntry, e routerEntry) []routerEntry {
	n := len(s)
	// Get the insert position
	// We are sorting all the patterns by len in descending order
	i := sort.Search(n, func(i int) bool {
		return len(e.pattern) > len(s[i].pattern)
	})
	s = append(s, routerEntry{})
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

func (r *Router) HandleFunc(pattern string, h func(ctx context.Context, event *ziggurat.Event) error) {
	if pattern == "" {
		panic(fmt.Errorf("pattern cannot be %q", pattern))
	}
	if h == nil {
		panic("handler cannot be <nil>")
	}
	r.register(pattern, ziggurat.HandlerFunc(h))
}

func (r *Router) register(pattern string, h ziggurat.Handler) {
	if r.handlerEntry == nil {
		r.handlerEntry = make(map[string]routerEntry)
	}

	//strip off trailing slash from the pattern
	if pattern[len(pattern)-1] == '/' {
		pattern = pattern[:len(pattern)-1]
	}

	if pattern == "" {
		panic(`"/" is not a valid pattern`)
	}

	//panic on multiple registrations
	if _, ok := r.handlerEntry[pattern]; ok {
		panic(fmt.Sprintf("multiple regirstrations for %s", pattern))
	}

	e := routerEntry{handler: h, pattern: pattern}

	r.handlerEntry[pattern] = e

	r.es = sortAndAppend(r.es, e)
}

func (r *Router) Handle(ctx context.Context, event *ziggurat.Event) error {
	path := event.RoutingPath
	h, _ := r.match(path)
	if h != nil {
		return h.Handle(ctx, event)
	}
	return fmt.Errorf("no pattern registered for %s", path)
}
