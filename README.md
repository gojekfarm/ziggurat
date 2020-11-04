### Ziggurat Golang
Stream Processing made easy

### Installation

- Run go get
```shell script
go get -v -u github.com/gojekfarm/ziggurat-go                                                                                                                                                       
```

#### How to use
- create a `config/config.yaml` in your project root
- sample `config.yaml`
```yaml
service-name: "test-app"
stream-router:
  message-log:
    bootstrap-servers: "localhost:9094"
    instance-count: 4
    # how many consumers to spawn.
    # adjust this number to the number of partitions
    # to maximize parallelization
    origin-topics: "^.*-message-log"
    group-id: "message_log_go"
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
If you wish to read config from a location other than the default location run the app with `--ziggurat-config="your_path/config_file.yaml"` option

```go
package main

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/zig"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type JSONMessage struct {
	Value int `json:"value"`
}

func main() {
	app := zig.NewApp()
	router := zig.NewRouter()

	router.HandlerFunc("plain-text-log", func(messageEvent zig.MessageEvent, a *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.MessageMetricsPublisher, zig.MessageLogger)

	router.HandlerFunc("json-log", func(messageEvent zig.MessageEvent, app *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.MessageLogger, zig.JSONDeserializer(JSONMessage{}))

	startFunc := func(a *zig.App) {
		fmt.Println("starting app")
	}

	stopFunc := func() {
		fmt.Println("stopping app")
	}

	<-app.Run(router, zig.RunOptions{
		HTTPConfigFunc: func(a *zig.App, r *httprouter.Router) {
			r.GET("/test_ping", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				writer.Write([]byte("TEST_PONG"))
			})
		},
		StartCallback: startFunc,
		StopCallback:  stopFunc,
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

## Known issues
- Data race occurs when rabbitmq is trying to re-connect, this doesn't lead to any adverse effects or message loss.
However this is not caused by ziggurat-go but by a library called `amqp-safe`.


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
- [ ] Unit tests
