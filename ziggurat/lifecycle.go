package ziggurat

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type LifeCycleHooks struct {
	StartFunction StartFunction
	StopFunction  StopFunction
}

func interruptHandler(interruptCh chan os.Signal, cancelFn context.CancelFunc, hooks LifeCycleHooks) {
	<-interruptCh
	log.Info().Msg("sigterm received, stopping consumers")
	cancelFn()
	hooks.StopFunction()
}

func Start(router *StreamRouter, lifeCycleHooks LifeCycleHooks) {
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	ctx, cancelFn := context.WithCancel(context.Background())

	go interruptHandler(interruptChan, cancelFn, lifeCycleHooks)

	parseConfig()
	log.Info().Msg("successfully parsed config")
	config := GetConfig()
	ConfigureLogger(config.LogLevel)
	lifeCycleHooks.StartFunction(config)
	router.Start(ctx, config)
}
