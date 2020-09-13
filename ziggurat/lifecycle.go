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
	HTTPServer    HttpServer
}

func interruptHandler(interruptCh chan os.Signal, cancelFn context.CancelFunc, appContext ApplicationContext, options *StartupOptions) {
	<-interruptCh
	log.Info().Msg("sigterm received")
	cancelFn()
	if err := appContext.Retrier.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping retrier")
	}

	if err := appContext.HttpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping http server")
	}

	options.StopFunction()
}

func initializeComponents(applicationContext *ApplicationContext, options *StartupOptions) {
	if options.Retrier == nil {
		applicationContext.Retrier = &RabbitRetrier{}
	}

	if options.HTTPServer == nil {
		applicationContext.HttpServer = &DefaultHttpServer{}
	}
}

func Start(router *StreamRouter, options StartupOptions) {
	interruptChan := make(chan os.Signal)
	applicationContext := ApplicationContext{}
	initializeComponents(&applicationContext, &options)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	ctx, cancelFn := context.WithCancel(context.Background())
	go interruptHandler(interruptChan, cancelFn, applicationContext, &options)

	parseConfig()
	log.Info().Msg("successfully parsed config")
	config := GetConfig()
	if validationErr := config.Validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("config validation error")
	}
	applicationContext.Config = config
	applicationContext.StreamRouter = router
	options.StartFunction(config)
	ConfigureLogger(config.LogLevel)
	<-router.Start(ctx, applicationContext)
}
