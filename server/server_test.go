package server

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"testing"
	"time"
)

const serverAddr = "localhost:8080"

func TestDefaultHttpServer_Start(t *testing.T) {
	ds := NewHTTPServer()
	ctx, cfn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cfn()
	errChan := ds.Run(ctx)
	time.Sleep(1 * time.Second)
	_, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Errorf("error connection to server %s", err)
	}
	<-errChan
}

func TestDefaultHttpServer_Stop(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cfn()
	ds := NewHTTPServer()
	errChan := ds.Run(ctx)
	<-errChan
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
