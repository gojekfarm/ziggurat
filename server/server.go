package server

import (
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var defaultHTTPPort = "8080"

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

func New(config ztype.ConfigStore) ztype.Server {
	port := config.Config().HTTPServer.Port
	if port == "" {
		port = defaultHTTPPort
	}
	router := httprouter.New()
	server := &http.Server{Addr: ":" + port, Handler: requestLogger(router)}
	return &DefaultHttpServer{
		server: server,
		router: router,
	}
}

func (s *DefaultHttpServer) Start(app ztype.App) error {
	s.router.POST("/v1/dead_set/:topic_entity/:count", replayHandler(app))
	s.router.GET("/v1/ping", pingHandler)

	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			zlog.LogError(err, "default http server error", nil)
		}
	}(s.server)
	return nil
}

func (s *DefaultHttpServer) ConfigureRoutes(a ztype.App, configFunc func(a ztype.App, h http.Handler)) {
	configFunc(a, s.router)
}

func (s *DefaultHttpServer) Handler() http.Handler {
	return s.router
}

func (s *DefaultHttpServer) Stop(app ztype.App) {
	zlog.LogError(s.server.Shutdown(app.Context()), "default http server: stopping http server", nil)
}
