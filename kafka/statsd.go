package kafka

import "github.com/cactus/go-statsd-client/v5/statsd"

type statsdClient struct {
	*statsd.Client
}
