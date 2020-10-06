package ziggurat

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

func (s *DefaultHttpServer) Start(ctx context.Context, app App) {
	router := httprouter.New()
	router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		count, _ := strconv.Atoi(params.ByName("count"))
		if replayErr := app.Retrier.Replay(app, params.ByName("topic_entity"), count); replayErr != nil {
			http.Error(writer, replayErr.Error(), 500)
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

	server := &http.Server{Addr: ":8080", Handler: router}
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			ServerLogger.Fatal().Err(err)
		}
	}(server)

	s.server = server

}

func (s *DefaultHttpServer) Stop() error {
	return s.server.Shutdown(context.Background())
}
