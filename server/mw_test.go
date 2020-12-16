package server

import (
	"github.com/gojekfarm/ziggurat"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequestLoggerMW(t *testing.T) {
	expectedArgs := map[string]interface{}{"METHOD": "GET", "URL": "/test"}
	oldLogger := ziggurat.LogInfo
	defer func() {
		ziggurat.LogInfo = oldLogger
	}()
	ziggurat.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(expectedArgs, args) {
			t.Errorf("expected %v got %v", expectedArgs, args)
		}
	}
	rl := requestLogger(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	rl.ServeHTTP(recorder, req)
}
