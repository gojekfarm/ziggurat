package server

import (
	"github.com/gojekfarm/ziggurat"
	"net/http"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		args := map[string]interface{}{"METHOD": request.Method, "URL": request.URL.Path}
		ziggurat.LogInfo("http server", args)
		next.ServeHTTP(writer, request)
	})
}
