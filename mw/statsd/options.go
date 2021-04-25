package statsd

import "github.com/gojekfarm/ziggurat"

type StatsDTag = map[string]string

// WithPrefix prepends a prefix to every every metric produced
func WithPrefix(prefix string) func(c *Client) {
	return func(c *Client) {
		c.prefix = prefix
	}
}

func WithHost(host string) func(c *Client) {
	return func(c *Client) {
		c.host = host
	}
}

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
