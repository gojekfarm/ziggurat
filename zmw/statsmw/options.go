package statsmw

import "github.com/gojekfarm/ziggurat/ztype"

func WithPrefix(prefix string) func(s *StatsDClient) {
	return func(s *StatsDClient) {
		s.prefix = prefix
	}
}

func WithHost(host string) func(s *StatsDClient) {
	return func(s *StatsDClient) {
		s.host = host
	}
}

func WithHandler(handler ztype.MessageHandler) func(s *StatsDClient) {
	return func(s *StatsDClient) {
		s.handler = handler
	}
}
