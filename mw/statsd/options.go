package statsd

import (
	"context"
	"github.com/gojekfarm/ziggurat/v2"
)

type Tags = map[string]string

type runOpts struct {
	GORoutinePublisher func(ctx context.Context)
}

// WithPrefix prepends a prefix to every metric produced
func WithPrefix(prefix string) func(c *Client) {
	return func(c *Client) {
		c.prefix = prefix
	}
}

// WithHost lets you specify a custom host
func WithHost(host string) func(c *Client) {
	return func(c *Client) {
		c.host = host
	}
}

// WithLogger lets you specify a custom `ziggurat.StructuredLogger`
func WithLogger(l ziggurat.StructuredLogger) func(c *Client) {
	return func(c *Client) {
		c.logger = l
	}
}

// WithDefaultTags sets global tags on every metric published.
// Could be a common app name,host name tag
func WithDefaultTags(tags Tags) func(c *Client) {
	return func(c *Client) {
		c.defaultTags = tags
	}
}
