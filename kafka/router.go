package kafka

import "github.com/gojekfarm/ziggurat"

/*
Example route <bootstrap_server>/<topic>/<parition>
entry {
	handler ziggurat.Handler
	pattern string
}
entries []string sorted by len of paths
*/

type entry struct {
	handler ziggurat.Handler
	pattern string
}

type Router struct {
	entries map[string]entry
	es      []string
}

func NewRouter() *Router {
	return &Router{
		entries: map[string]entry{},
	}
}
