package rabbitmq

type QueueConfig struct {
	QueueKey              string
	DelayExpirationInMS   string
	RetryCount            int
	ConsumerPrefetchCount int
	ConsumerCount         int
}

type Queues []QueueConfig
