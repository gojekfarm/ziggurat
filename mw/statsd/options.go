package statsd

import "github.com/gojekfarm/ziggurat"

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
		c.Logger = l
	}
}
