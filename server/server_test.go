package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

const serverAddr = "localhost:8080"

func TestDefaultHttpServer_Start(t *testing.T) {
	ds := NewHTTPServer()
	ctx, cfn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cfn()
	go func() { ds.Run(ctx) }()
	_, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Errorf("error connection to server %s", err)
	}

}

func TestDefaultHttpServer_Stop(t *testing.T) {
	ctx, cfn := context.WithCancel(context.Background())

	ds := NewHTTPServer()
	go func() { ds.Run(ctx) }()
	time.Sleep(500 * time.Millisecond)
	cfn()
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
