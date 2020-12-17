package server

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"testing"
	"time"
)

const serverAddr = "localhost:8080"

func TestDefaultHttpServer_Start(t *testing.T) {
	a := ziggurat.NewZig()
	ds := NewHTTPServer()
	ds.Start(a)
	time.Sleep(100 * time.Millisecond)
	_, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Errorf("error connection to server %s", err)
	}
	ds.Stop(a)
}

func TestDefaultHttpServer_Stop(t *testing.T) {
	a := ziggurat.NewZig()
	ds := NewHTTPServer()
	ds.Start(a)
	ds.Stop(a)
	if _, err := net.Dial("tcp", serverAddr); err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestDefaultHttpServer_ConfigureHTTPRoutes(t *testing.T) {
	ds := NewHTTPServer()
	ds.ConfigureHTTPEndpoints(func(r *httprouter.Router) {
		r.GET("/test_route", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})
	})
	if handle, _, _ := ds.router.Lookup(http.MethodGet, "/test_route"); handle == nil {
		t.Errorf("failed to attach route")
	}
}
