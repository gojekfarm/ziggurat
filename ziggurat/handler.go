package zig

import (
	"github.com/rs/zerolog/log"
	"time"
)

func MessageHandler(app *App, handlerFunc HandlerFunc) func(event MessageEvent) {
	return func(event MessageEvent) {
		metricTags := map[string]string{
			"topic_entity": event.TopicEntity,
			"kafka_topic":  event.Topic,
		}
		funcExecStartTime := time.Now()
		status := handlerFunc(event, app)
		funcExecEndTime := time.Now()
		app.metricPublisher.Gauge("handler_func_exec_time", funcExecEndTime.Sub(funcExecStartTime).Milliseconds(), metricTags)
		switch status {
		case ProcessingSuccess:
			if publishErr := app.metricPublisher.IncCounter("message_processing_success", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("[ZIG MESSAGE HANDLER]")
			}
			log.Info().Msg("[ZIG MESSAGE HANDLER] successfully processed message")
		case SkipMessage:
			if publishErr := app.metricPublisher.IncCounter("message_processing_failure_skip", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			log.Info().Msg("[ZIG MESSAGE HANDLER] skipping message")

		case RetryMessage:
			log.Info().Msg("[ZIG MESSAGE HANDLER] retrying message")
			if publishErr := app.metricPublisher.IncCounter("message_processing_failure_skip", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			if retryErr := app.messageRetry.Retry(app, event); retryErr != nil {
				log.Error().Err(retryErr).Msg("[ZIG MESSAGE HANDLER] error retrying message")
			}
		default:
			log.Error().Err(ErrInvalidReturnCode).Msg("[ZIG MESSAGE HANDLER] return code must be one of `zig.ProcessingSuccess OR zig.RetryMessage OR zig.SkipMessage`")
		}
	}
}
