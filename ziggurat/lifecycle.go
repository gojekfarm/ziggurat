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

func interruptHandler(interruptCh chan os.Signal, cancelFn context.CancelFunc, options *StartupOptions) {
	<-interruptCh
	log.Info().Msg("sigterm received")
	cancelFn()
	if err := options.Retrier.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping retrier")
	}
	options.StopFunction()
}

func Start(router *StreamRouter, options StartupOptions) {
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	ctx, cancelFn := context.WithCancel(context.Background())
	go interruptHandler(interruptChan, cancelFn, &options)

	if options.Retrier == nil {
		options.Retrier = &RabbitRetrier{}
	}

	parseConfig()
	log.Info().Msg("successfully parsed config")
	config := GetConfig()
	if validationErr := config.Validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("config validation error")
	}

	ConfigureLogger(config.LogLevel)
	<-router.Start(ctx, config, options.Retrier)
}
