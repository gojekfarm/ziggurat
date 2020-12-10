package server

import (
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"testing"
	"time"
)

const serverAddr = "localhost:8080"

func TestDefaultHttpServer_Start(t *testing.T) {
	cs := mock.NewConfigStore()
	a := mock.NewApp()
	a.ConfigStoreFunc = func() ztype.ConfigStore {
		return cs
	}
	cs.ConfigFunc = func() *zbase.Config {
		return &zbase.Config{
			HTTPServer: zbase.HTTPServerConfig{Port: "8080"},
		}
	}
	ds := New(cs)
	ds.Start(a)
	time.Sleep(5 * time.Second)
	_, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Errorf("error connection to server %s", err)
	}
	ds.Stop(a)
}

func TestDefaultHttpServer_Stop(t *testing.T) {
	cs := mock.NewConfigStore()
	ds := New(cs)
	a := mock.NewApp()
	ds.Start(a)
	ds.Stop(a)
	if _, err := net.Dial("tcp", serverAddr); err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestDefaultHttpServer_ConfigureHTTPRoutes(t *testing.T) {
	cs := mock.NewConfigStore()
	a := mock.NewApp()
	ds := New(cs)
	ds.Start(a)
	ds.ConfigureRoutes(a, func(a ztype.App, h http.Handler) {
		r := h.(*httprouter.Router)
		r.GET("/test_route", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})
	})
	r := ds.Handler().(*httprouter.Router)
	if handle, _, _ := r.Lookup(http.MethodGet, "/test_route"); handle == nil {
		t.Errorf("failed to attach route")
	}
}
