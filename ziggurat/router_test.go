package ziggurat

import "testing"

var dr = NewRouter()

func TestDefaultRouter_HandlerFunc(t *testing.T) {
	topicEntity := "test-entity"
	topicEntityTwo := "test-entity2"
	dr.HandlerFunc(topicEntity, func(messageEvent MessageEvent, app *Ziggurat) ProcessStatus {
		return ProcessingSuccess
	})
	dr.HandlerFunc(topicEntityTwo, func(messageEvent MessageEvent, app *Ziggurat) ProcessStatus {
		return ProcessingSuccess
	})
	if len(dr.handlerFunctionMap) < 2 {
		t.Errorf("expected %d entries in handlerFunctionMap but got %d", 2, len(dr.handlerFunctionMap))
	}
}
