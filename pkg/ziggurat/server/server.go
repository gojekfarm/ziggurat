package server

import (
	"context"
	"encoding/json"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/z"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

var defaultHTTPPort = "8080"

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

type ReplayResponse struct {
	Status bool   `json:"status"`
	Count  int    `json:"count"`
	Msg    string `json:"msg"`
}

func NewDefaultHTTPServer(config z.ConfigReader) z.HttpServer {
	port := config.Config().HTTPServer.Port
	if port == "" {
		port = defaultHTTPPort
	}
	router := httprouter.New()
	server := &http.Server{Addr: ":" + port, Handler: router}
	return &DefaultHttpServer{
		server: server,
		router: router,
	}
}

func (s *DefaultHttpServer) Start(app z.App) {
	s.router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		count, _ := strconv.Atoi(params.ByName("count"))
		if replayErr := app.MessageRetry().Replay(app, params.ByName("topic_entity"), count); replayErr != nil {
			http.Error(writer, replayErr.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		jsonBytes, _ := json.Marshal(ReplayResponse{
			Status: true,
			Count:  count,
			Msg:    "successfully replayed messages",
		})
		writer.Header().Add("Content-Type", "application/json")
		writer.Write(jsonBytes)
	})

	s.router.GET("/v1/ping", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("pong"))
	})

	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			logger.LogError(err, "ziggurat http-server:", nil)
		}
	}(s.server)

}

func (s *DefaultHttpServer) ConfigureHTTPRoutes(a z.App, configFunc func(a z.App, h http.Handler)) {
	configFunc(a, s.router)
}

func (s *DefaultHttpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
