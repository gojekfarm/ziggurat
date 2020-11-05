package main

import (
	"github.com/gojekfarm/ziggurat-go/zig"
)

func main() {
	app := zig.NewApp()
	router := zig.NewRouter()

	router.HandlerFunc("{{.TopicEntity}}", func(messageEvent zig.MessageEvent, app *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.MessageLogger)

	<-app.Run(router, zig.RunOptions{
		StartCallback: func(a *zig.App) {

		},
		StopCallback: func() {

		},
	})

}