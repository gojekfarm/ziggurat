package statsd

import (
	"github.com/gojekfarm/ziggurat"
)

func WithPrefix(prefix string) func(s *Client) {
	return func(s *Client) {
		s.prefix = prefix
	}
}

func WithHost(host string) func(s *Client) {
	return func(s *Client) {
		s.host = host
	}
}

func WithHandler(handler ziggurat.Handler) func(s *Client) {
	return func(s *Client) {
		s.handler = handler
	}
}
