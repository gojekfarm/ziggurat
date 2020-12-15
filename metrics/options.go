package metrics

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
