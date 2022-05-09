//go:build ignore
// +build ignore

package main

import (
    "context"
    "github.com/gojekfarm/ziggurat/mw/statsd"
    "strconv"
    "strings"
    "time"

    "github.com/gojekfarm/ziggurat"
    "github.com/gojekfarm/ziggurat/kafka"
    "github.com/gojekfarm/ziggurat/logger"
    "github.com/gojekfarm/ziggurat/mw/rabbitmq"
)

func main() {
    var zig ziggurat.Ziggurat
    var r kafka.Router

    ctx := context.Background()
    l := logger.NewLogger(logger.LevelInfo)

    statsClient := statsd.NewPublisher(statsd.WithLogger(l),
        statsd.WithDefaultTags(statsd.StatsDTag{"ziggurat-version": "v162"}),
        statsd.WithPrefix("ziggurat_v162"))

    ks := kafka.Streams{
        StreamConfig: kafka.StreamConfig{{
            BootstrapServers: "localhost:9092",
            Topics:           "plain-text-log",
            GroupID:          "ziggurat_consumer",
            ConsumerCount:    2,
            RouteGroup:       "plain-text-messages"}},
        Logger: l,
    }

    ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{
        {
            QueueKey:              "plain_text_messages_retry",
            DelayExpirationInMS:   "500",
            ConsumerPrefetchCount: 5,
            ConsumerCount:         100,
            RetryCount:            2,
        },
    }, rabbitmq.WithLogger(l),
        rabbitmq.WithUsername("user"),
        rabbitmq.WithPassword("bitnami"),
        rabbitmq.WithConnectionTimeout(3*time.Second))

    r.HandleFunc("plain-text-messages/", func(ctx context.Context, event *ziggurat.Event) error {
        val := string(event.Value)
        s := strings.Split(val, "_")
        num, err := strconv.Atoi(s[1])
        if err != nil {
            return err
        }
        if num%2 == 0 {
            return ar.Retry(ctx, event, "plain_text_messages_retry")
        }
        return nil
    })

    h := ziggurat.Use(&r, statsClient.PublishEventDelay, statsClient.PublishHandlerMetrics)

    if runErr := zig.RunAll(ctx, h, &ks, ar); runErr != nil {
        l.Error("error running streams", runErr)
    }

}
