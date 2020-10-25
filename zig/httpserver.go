package zig

import (
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

type ReplayResponse struct {
	Status bool   `json:"status"`
	Count  int    `json:"count"`
	Msg    string `json:"msg"`
}

func NewDefaultHTTPServer(config *Config) HttpServer {
	port := config.HTTPServer.Port
	router := httprouter.New()
	server := &http.Server{Addr: ":" + port, Handler: router}
	return &DefaultHttpServer{
		server: server,
		router: router,
	}
}

func (s *DefaultHttpServer) Start(app *App) {
	s.router = httprouter.New()
	s.router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		count, _ := strconv.Atoi(params.ByName("count"))
		if replayErr := app.retrier.Replay(app, params.ByName("topic_entity"), count); replayErr != nil {
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
			serverLogger.Error().Err(err)
		}
	}(s.server)

}

func (s *DefaultHttpServer) DefineRoutes(attachFunc func(r *httprouter.Router)) {
	attachFunc(s.router)
}

func (s *DefaultHttpServer) Stop() error {
	return s.server.Shutdown(context.Background())
}
