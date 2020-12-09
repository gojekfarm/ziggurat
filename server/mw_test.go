package server

import (
	"github.com/gojekfarm/ziggurat/zlog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequestLoggerMW(t *testing.T) {
	expectedArgs := map[string]interface{}{"METHOD": "GET", "URL": "/test"}
	zlog.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(expectedArgs, args) {
			t.Errorf("expected args %v got %v", expectedArgs, args)
		}
	}
	rl := requestLogger(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	rl.ServeHTTP(recorder, req)
}
