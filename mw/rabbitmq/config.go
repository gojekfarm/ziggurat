package rabbitmq

type QueueConfig struct {
	QueueName             string
	DelayExpirationInMS   string
	RetryCount            int
	ConsumerPrefetchCount int
	ConsumerCount         int
}

type Queues []QueueConfig
