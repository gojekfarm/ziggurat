package server

import (
	"github.com/gojekfarm/ziggurat"
	"net/http"
)

func HTTPRequestLogger(logger ziggurat.StructuredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			logger.Info("PATH=%s METHOD=%s", request.URL.String(), request.Method)
			next.ServeHTTP(writer, request)
		})
	}
}
