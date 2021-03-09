package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gojekfarm/ziggurat/logger"
)

func TestRequestLoggerMW(t *testing.T) {
	expectedArgs := []map[string]interface{}{{"path": "/test", "method": http.MethodGet}}
	mockLogger := logger.DiscardLogger{}
	mockLogger.InfoFunc = func(m string, kv ...map[string]interface{}) {
		if !reflect.DeepEqual(kv, expectedArgs) {
			t.Errorf("expected %+v got %+v", expectedArgs, kv)
		}
	}
	requestLogger := HTTPRequestLogger(mockLogger)
	rl := requestLogger(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	rl.ServeHTTP(recorder, req)
}
