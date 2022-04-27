package main

import (
    "context"
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

    ks := kafka.Streams{
        StreamConfig: kafka.StreamConfig{{
            BootstrapServers: "g-gojek-id-mainstream.golabs.io:6668",
            Topics:           "driver-location-ping-3",
            GroupID:          "dlr_pings_go_ziggurat_02",
            ConsumerCount:    2,
            RouteGroup:       "dlr_ping"}},
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

    r.HandleFunc("plain-text-messages/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
        val := string(event.Value)
        s := strings.Split(val, "_")
        num, err := strconv.Atoi(s[1])
        if err != nil {
            return err
        }
        if num%2 == 0 {
            return ziggurat.Retry
        }
        return nil
    }, "plain_text_messages_retry"))

    h := ziggurat.Use(&r)

    if runErr := zig.RunAll(ctx, h, &ks, ar); runErr != nil {
        l.Error("error running streams", runErr)
    }

}
