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
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/stream"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zig"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func main() {
	app := zig.NewApp()
	router := stream.NewRouter()

	router.HandlerFunc("plain-text-log", func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})

	startFunc := func(a z.App) {
		fmt.Println("starting app")
	}

	stopFunc := func() {
		fmt.Println("stopping app")
	}

	<-app.Run(router, z.RunOptions{
		StartCallback: startFunc,
		StopCallback:  stopFunc,
		HTTPConfigFunc: func(a z.App, h http.Handler) {
			r, _ := h.(*httprouter.Router)
			r.GET("/if_you_dont_eat_your_meat", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				writer.Write([]byte("you can't have any pudding"))
			})
		},
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
- <Data race occurs when rabbitmq is trying to re-connect, this doesn't lead to any adverse effects or message loss.
However this is not caused by ziggurat-go but by a library called `amqp-safe`. [Fixed]


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
