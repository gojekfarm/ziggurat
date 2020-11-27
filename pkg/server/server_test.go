package server

import (
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"testing"
	"time"
)

const serverAddr = "localhost:8080"

func TestDefaultHttpServer_Start(t *testing.T) {
	configStore := mock.NewConfigStore()
	a := mock.NewApp()
	a.ConfigFunc = func() *zbasic.Config {
		return &zbasic.Config{HTTPServer: zbasic.HTTPServerConfig{Port: "400"}}
	}
	ds := NewDefaultHTTPServer(configStore)
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
	ds := NewDefaultHTTPServer(cs)
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
	ds := NewDefaultHTTPServer(cs)
	ds.Start(a)
	ds.ConfigureRoutes(a, func(a z.App, h http.Handler) {
		r := h.(*httprouter.Router)
		r.GET("/test_route", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})
	})
	r := ds.Handler().(*httprouter.Router)
	if handle, _, _ := r.Lookup(http.MethodGet, "/test_route"); handle == nil {
		t.Errorf("failed to attach route")
	}
}
