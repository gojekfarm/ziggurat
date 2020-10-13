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
}

type ReplayResponse struct {
	Status bool   `json:"status"`
	Count  int    `json:"status"`
	Err    string `json:"error"`
}

func (s *DefaultHttpServer) Start(app *App) {
	router := httprouter.New()
	port := app.Config.HTTPServer.Port
	router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		count, _ := strconv.Atoi(params.ByName("count"))
		if replayErr := app.Retrier.Replay(app, params.ByName("topic_entity"), count); replayErr != nil {
			http.Error(writer, replayErr.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(200)
		jsonBytes, _ := json.Marshal(ReplayResponse{
			Status: true,
			Count:  count,
			Err:    "",
		})
		writer.Header().Add("Content-Type", "application/json")
		writer.Write(jsonBytes)
	})

	router.GET("/v1/ping", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("pong"))
	})

	server := &http.Server{Addr: ":" + port, Handler: router}
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			serverLogger.Fatal().Err(err)
		}
	}(server)

	s.server = server

}

func (s *DefaultHttpServer) Stop() error {
	return s.server.Shutdown(context.Background())
}
