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

- To override the boostrap-servers under plain-text-log

```shell script
export ZIGGURAT_STREAM_ROUTER_PLAIN_TEXT_LOG_BOOTSTRAP_SERVERS="localhost:9094"
```

- sample `main.go`

### Pass an optional config file path

If you wish to read config from a location other than the default location run the app
with `--ziggurat-config="your_path/config_file.yaml"` option

```go
package main

import (
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/gojekfarm/ziggurat/zapp"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zmw"
	"github.com/gojekfarm/ziggurat/zrouter"
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