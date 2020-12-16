package server

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
