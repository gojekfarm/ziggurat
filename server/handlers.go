package server

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type ReplayResponse struct {
	Status bool   `json:"status"`
	Count  int    `json:"count"`
	Msg    string `json:"msg"`
}

func pingHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("pong"))
}
