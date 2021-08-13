package kafka

import (
	"context"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

/*ConsumerConfig takes in the kafka consumer config
ConsumerConfig {
	BootstrapServers "localhost:9092,localhost:9093"
	OriginTopics     "song-likes"
	ConsumerGroupID  "song_likes_consumer"
	ConsumerCount    4
	RouteGroup       "song-likes"
}*/
type ConsumerConfig struct {
	BootstrapServers string
	OriginTopics     string
	ConsumerGroupID  string
	ConsumerCount    int
	RouteGroup       string
}

type StreamConfig = []ConsumerConfig

type Streams struct {
	routeConsumerMap map[string][]*kafka.Consumer
	Logger           ziggurat.StructuredLogger
	StreamConfig     StreamConfig
}

// Stream implements the ziggurat.Streamer interface
// blocks until an error is received or context is cancelled
func (k *Streams) Stream(ctx context.Context, handler ziggurat.Handler) error {
	if k.Logger == nil {
		k.Logger = logger.NewJSONLogger("info")
	}
	var wg sync.WaitGroup
	k.routeConsumerMap = make(map[string][]*kafka.Consumer, len(k.StreamConfig))
	for _, stream := range k.StreamConfig {
		consumerConfig := NewConsumerConfig(stream.BootstrapServers, stream.ConsumerGroupID)
		topics := strings.Split(stream.OriginTopics, ",")
		routeName := stream.RouteGroup
		k.routeConsumerMap[routeName] = StartConsumers(ctx, consumerConfig, routeName, topics, stream.ConsumerCount, handler, k.Logger, &wg)
	}

	wg.Wait()
	k.stop()

	return nil
}

func (k *Streams) stop() {
	for _, consumers := range k.routeConsumerMap {
		for i := range consumers {
			k.Logger.Error("error stopping consumer %v", consumers[i].Close(), nil)
		}
	}
}
