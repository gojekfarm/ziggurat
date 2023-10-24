package rabbitmq

const (
	RetryQueue = iota
	WorkerQueue
)

type QueueConfig struct {
	QueueKey              string
	DelayExpirationInMS   string
	RetryCount            int
	ConsumerPrefetchCount int
	ConsumerCount         int
	Type                  int
	DiscardCorruptMessage bool
}

type Queues []QueueConfig
