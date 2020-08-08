package ziggurat

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type StartupOptions struct {
	StartFunction StartFunction
	StopFunction  StopFunction
	Retrier       MessageRetrier
}

func interruptHandler(interruptCh chan os.Signal, cancelFn context.CancelFunc, stopFunction StopFunction) {
	<-interruptCh
	log.Info().Msg("sigterm received")
	cancelFn()
	stopFunction()
}

func Start(router *StreamRouter, options StartupOptions) {
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	ctx, cancelFn := context.WithCancel(context.Background())

	go interruptHandler(interruptChan, cancelFn, options.StopFunction)

	parseConfig()
	log.Info().Msg("successfully parsed config")
	config := GetConfig()
	ConfigureLogger(config.LogLevel)
	options.StartFunction(config)
	<-router.Start(ctx, config)
}
