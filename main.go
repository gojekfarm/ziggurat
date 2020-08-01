package main

import (
	"fmt"
	"github.com/gojek/ziggurat"
)

func main() {
	sr := ziggurat.NewStreamRouter()
	sr.HandlerFunc("test-entity", func(message interface{}) {
		fmt.Printf("[handlerFunc]: Received message for test-entity1 %v\n", message)
	})

	sr.HandlerFunc("test-entity2", func(message interface{}) {
		fmt.Printf("[handlerFunc]: Received message for test-entity2 %v\n", message)
	})

	ziggurat.Start(sr, ziggurat.LifeCycleHooks{
		StartFunction: func(config ziggurat.ZigguratConfig) {
			fmt.Printf("Starting app...\n")
		},
		StopFunction: func(config ziggurat.ZigguratConfig) {
			fmt.Printf("Stopping app...\n")
		},
	})

}
