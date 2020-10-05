package ziggurat

import (
	"github.com/rs/zerolog/log"
)

func MessageHandler(app App, handlerFunc HandlerFunc, retrier MessageRetrier) func(event MessageEvent) {
	return func(event MessageEvent) {
		metricTags := map[string]string{
			"topic_entity": event.TopicEntity,
			"kafka_topic":  event.Topic,
		}
		status := handlerFunc(event)
		switch status {
		case ProcessingSuccess:
			if publishErr := app.MetricPublisher.IncCounter("message_processing_success", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}

			log.Info().Msg("successfully processed message")
		case SkipMessage:
			log.Info().Msg("skipping message")

		case RetryMessage:
			log.Info().Msgf("retrying message")
			if publishErr := app.MetricPublisher.IncCounter("message_processing_failure", 1, metricTags); publishErr != nil {
				log.Error().Err(publishErr).Msg("")
			}
			if retryErr := retrier.Retry(app, event); retryErr != nil {
				log.Error().Err(retryErr).Msg("error retrying message")
			}
		}
	}
}
