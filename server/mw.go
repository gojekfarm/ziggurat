package server

import (
	"net/http"

	"github.com/gojekfarm/ziggurat"
)

func HTTPRequestLogger(logger ziggurat.StructuredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(writer http.ResponseWriter, request *http.Request) {
			kvs := map[string]interface{}{"path": request.URL.String(), "method": request.Method}
			logger.Info("http request logger", kvs)
			next.ServeHTTP(writer, request)
		}
		return http.HandlerFunc(f)
	}
}
