package server

import (
	"github.com/gojekfarm/ziggurat"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerMW(t *testing.T) {
	logger := ziggurat.NewMockLogger("info")
	logger.InfofFunc = func(format string, v ...interface{}) {

	}
	requestLogger := HTTPRequestLogger(ziggurat.NewLogger("info"))
	rl := requestLogger(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	rl.ServeHTTP(recorder, req)
}
