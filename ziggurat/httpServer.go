package ziggurat

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type HttpServer struct{}

func (s *HttpServer) Start(config Config, handlerMap TopicEntityHandlerMap) error {
	router := httprouter.New()
	router.POST("/v1/dead_set/:topic_entity/:count", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	})
	serverErr := http.ListenAndServe(":8080", router)
	return serverErr
}

func (s *HttpServer) Stop() error {
	return nil
}
