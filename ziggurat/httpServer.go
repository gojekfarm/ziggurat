package ziggurat

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type DefaultHttpServer struct {
	server  *http.Server
	retrier MessageRetrier
}

func (s *DefaultHttpServer) Start(ctx context.Context, applicationContext App) {
	router := httprouter.New()
	router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		count, _ := strconv.Atoi(params.ByName("count"))
		if replayErr := applicationContext.Retrier.Replay(applicationContext, params.ByName("topic_entity"), count); replayErr != nil {
			http.Error(writer, replayErr.Error(), 500)
			return
		}
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "succesfully replayed %d messages", count)
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
