package ziggurat

import "github.com/rs/zerolog/log"

func MessageHandler(config Config, handlerFunc HandlerFunc, retrier MessageRetrier) func(event MessageEvent) {
	return func(event MessageEvent) {
		status := handlerFunc(event)
		switch status {
		case ProcessingSuccess:
			log.Info().Msgf("successfully processed message %v", event)
		case SkipMessage:
			log.Info().Msgf("skipping message %v", event)
		case RetryMessage:
			log.Info().Msgf("retrying message %v", event)
			if retryErr := retrier.Retry(config, event); retryErr != nil {
				log.Error().Err(retryErr).Msg("error retrying message")
			}
		}
	}
}
