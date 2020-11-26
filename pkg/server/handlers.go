package server

import (
	"encoding/json"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func replayHandler(app z.App) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
	}
}

func pingHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("pong"))
}
