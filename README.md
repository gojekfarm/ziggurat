### Ziggurat Golang
A stream processing framework in golang

### Installation

- Run go get
```shell script
go get -v -u github.com/gojekfarm/ziggurat-go                                                                                                                                                       
```

#### How to use
- create a `config/config.yaml` in your project root
- sample `config.yaml`
```yaml
service-name: "test-service"
stream-router:
  test-entity2:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic1"
    group-id: "test-group"
  test-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic2"
    group-id: "test-group2"
  json-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 1
    origin-topics: "json-test"
    group-id: "json-group"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
```
#### Overriding the config using ENV variables
- To override the boostrap-servers under test-entity2
```shell script
export ZIGGURAT_STREAM_ROUTER_TEST_ENTITY2_BOOTSTRAP_SERVERS="localhost:9094"
```


- sample `main.go`
### Pass an optional config file path
If you wish to read config from a location other than the default location run the app with `--config="your_path/config_file.yaml"` option

```go
package main

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/zig"
)

type JSONMessage struct {
	Value int `json:"value"`
}

func main() {
	app := zig.NewApp()
	router := zig.NewRouter()

	router.HandlerFunc("test-entity", func(messageEvent zig.MessageEvent, a *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.Middleware{zig.MessageLogger})

	router.HandlerFunc("json-entity", func(messageEvent zig.MessageEvent, app *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.Middleware{zig.MessageLogger, zig.JSONDeserializer(&JSONMessage{})})

	app.Configure(zig.Options{
		StopFunc: func() {
			fmt.Println("stopping app...")
		},
	})

	app.Run(router, func(a *zig.App) {
		fmt.Println("starting app...")
	})
}

```

### Using middlewares
Ziggurat allows you to use middlewares to add pluggable functionality to your stream handler logic,

How do I write my own middleware? 

Middlewares are just normal functions of type `zig.MiddlewareFunc` 

Example middleware

```go
package mypkg

import (
	"fmt"
)

func Logger(handlerFunc HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *App) ProcessStatus {
		fmt.Println("[HELLO FROM LOGGER]: Hello from logger")
		return handlerFunc(messageEvent, app)
	}
}
```
Ziggurat provides  middlewares for logging and de-serializing kafka messages


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
- [ ] Publish metrics
- [ ] Producer API
- [ ] Unit tests
