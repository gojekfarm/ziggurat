package kafka

import (
	"strings"

	"github.com/gojekfarm/ziggurat"
)

/*
Example route <bootstrap_server>/<topic>/<partition>
entry {
	handler ziggurat.Handler
	pattern string
}
entries []string sorted by len of paths
*/

//entry contains the pattern and the path entry
type entry struct {
	handler ziggurat.Handler
	pattern string
}

type Router struct {
	entries map[string]entry
	es      []entry
}

func NewRouter() *Router {
	return &Router{
		entries: map[string]entry{},
	}
}

//match works by matching the shortest prefix that matches the path
// it returns the matched path and the handler associated with it
func (r *Router) match(path string) (ziggurat.Handler, string) {
	if e, ok := r.entries[path]; ok {
		return e.handler, path
	}
	for _, e := range r.es {
		if strings.HasPrefix(path, e.pattern) {
			return e.handler, e.pattern
		}
	}

	return nil, ""
}
