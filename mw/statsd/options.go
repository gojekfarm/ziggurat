package statsd

import (
	"github.com/gojekfarm/ziggurat"
	"time"
)

type StatsDTag = map[string]string

type runOpts struct {
	goPublishInterval time.Duration
}

// WithPrefix prepends a prefix to every every metric produced
func WithPrefix(prefix string) func(c *Client) {
	return func(c *Client) {
		c.prefix = prefix
	}
}

//WithHost lets you specify a custom host
func WithHost(host string) func(c *Client) {
	return func(c *Client) {
		c.host = host
	}
}

//WithLogger lets you specify a custom `ziggurat.StructuredLogger`
func WithLogger(l ziggurat.StructuredLogger) func(c *Client) {
	return func(c *Client) {
		c.logger = l
	}
}

// WithDefaultTags sets global tags on every metric published.
// Could be a common app name,host name tag
func WithDefaultTags(tags StatsDTag) func(c *Client) {
	return func(c *Client) {
		c.defaultTags = tags
	}
}

// WithGoRoutinePublishInterval lets you specify the
// interval for publishing go-routine count
func WithGoRoutinePublishInterval(d time.Duration) func(r *runOpts) {
	return func(r *runOpts) {
		r.goPublishInterval = d
	}
}
