package server

import (
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplayHandler(t *testing.T) {
	app := mock.NewApp()
	retry := mock.NewRetry()
	app.MessageRetryFunc = func() z.MessageRetry {
		return retry
	}
	retry.ReplayFunc = func(app z.App, topicEntity string, count int) error {
		if topicEntity != "test" && count != 5 {
			t.Errorf("expected [test,5] got [%s,%d]", topicEntity, count)
		}
		return nil
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/dead_set/test/5", nil)
	r := httprouter.New()
	r.POST("/v1/dead_set/:topic_entity/:count", replayHandler(app))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	expectedJSON := `{"status":true,"count":5,msg:"successfully replayed messages"}`
	if rr.Code != 200 && rr.Body.String() != expectedJSON {
		t.Errorf("expected [200,%s] got [%d,%s]", expectedJSON, rr.Code, rr.Body.String())
	}
}

func TestReplayHandlerError(t *testing.T) {
	app := mock.NewApp()
	retry := mock.NewRetry()
	app.MessageRetryFunc = func() z.MessageRetry {
		return retry
	}
	retry.ReplayFunc = func(app z.App, topicEntity string, count int) error {
		return errors.New("replay error")
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/dead_set/test/5", nil)
	r := httprouter.New()
	r.POST("/v1/dead_set/:topic_entity/:count", replayHandler(app))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	expectedBody := "replay error"
	expectedStatus := 500
	if rr.Code != expectedStatus && rr.Body.String() != expectedBody {
		t.Errorf("expected [%d,%s] got [%d,%s]", expectedStatus, expectedBody, rr.Code, rr.Body.String())
	}
}

func TestPingHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	r := httprouter.New()
	r.GET("/v1/ping", pingHandler)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	expectedStatus := 200
	if rr.Code != expectedStatus && rr.Body.String() != "pong" {
		t.Errorf("expected [%d,%s], got [%d,%s]", expectedStatus, "pong", rr.Code, rr.Body.String())
	}
}
