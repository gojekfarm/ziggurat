### Ziggurat Golang

Stream Processing made easy

### Install the zig CLI

```shell script
go get -v -u github.com/gojekfarm/ziggurat-go/cmd/...
go install github.com/gojekfarm/ziggurat-go/cmd/...                                                                                                                                                       
```

#### How to use

- create a new app using the `new` command

```shell
zig new <app_name>
```

#### sample config file

```yaml
service-name: "test-app"
stream-router:
  plain-text-log:
  bootstrap-servers: "localhost:9092"
  instance-count: 2
  origin-topics: "plain-text-log"
  group-id: "plain_text_consumer"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
statsd:
  host: "localhost:8125"
http-server:
  port: 8080
```

#### Overriding the config using ENV variables

- To override the boostrap-servers under test-entity2

```shell script
export ZIGGURAT_STREAM_ROUTER_TEST_ENTITY2_BOOTSTRAP_SERVERS="localhost:9094"
```

- sample `main.go`

### Pass an optional config file path

If you wish to read config from a location other than the default location run the app
with `--ziggurat-config="your_path/config_file.yaml"` option

```go
package main

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/za"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"github.com/gojekfarm/ziggurat-go/pkg/zmw"
	"github.com/gojekfarm/ziggurat-go/pkg/zr"
)

func main() {
	app := za.NewApp()
	router := zr.NewRouter()

	router.HandleFunc("plain-text-log", func(event zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})

	routerWithMW := router.Compose(zmw.MessageLogger, zmw.MessageMetricsPublisher)

	<-app.Run(routerWithMW, router.Routes())
}
```

#### TODO

- [x] Balanced Consumer groups
- [x] RabbitMQ retries
- [x] At least once delivery semantics
- [x] Retry interface
- [x] Default middleware to deserialize messages
- [x] Env vars Config override
- [x] HTTP server
- [x] Replay RabbitMQ deadset messages
- [x] Log formatting
- [x] StatsD support
- [ ] Producer API
