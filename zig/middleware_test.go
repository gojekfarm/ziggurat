package zig

import (
	testproto "github.com/gojekfarm/ziggurat-go/protobuf"
	"github.com/golang/protobuf/proto"
	"testing"
)

type JSONMessage struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SecretNum int    `json:"secretNum"`
}

func TestJSONDeserializer(t *testing.T) {
	jsonMessage := `{"id":"xyzzyspoonshift1","name":"road_rash","secretNum":1}`
	jsonDeserializer := JSONDeserializer(JSONMessage{})
	expectedMessage := JSONMessage{
		ID:        "xyzzyspoonshift1",
		Name:      "road_rash",
		SecretNum: 1,
	}
	handler := func(event MessageEvent, app *Ziggurat) ProcessStatus {
		message := *event.MessageValue.(*JSONMessage)
		if message == expectedMessage {
			return ProcessingSuccess
		}
		return SkipMessage
	}
	result := jsonDeserializer(handler)(MessageEvent{
		MessageValueBytes: []byte(jsonMessage),
	}, &Ziggurat{})

	if result == SkipMessage {
		t.Errorf("expected %v but got %v", ProcessingSuccess, result)
	}
}

func TestProtobufDeserializer(t *testing.T) {
	testProtoModel := testproto.TestMessage{}
	expectedMessage := testproto.TestMessage{
		Id:        "1",
		SecretNum: 0,
		Name:      "1",
	}
	bytes, _ := proto.Marshal(&expectedMessage)
	//go vet complains copying lockers by value
	protoDeserializer := ProtobufDeserializer(testProtoModel)
	protoDeserializer(func(messageEvent MessageEvent, app *Ziggurat) ProcessStatus {
		tm := messageEvent.MessageValue.(*testproto.TestMessage)
		if !proto.Equal(tm, &expectedMessage) {
			t.Errorf("proto messages are not equal")
		}
		return ProcessingSuccess
	})(MessageEvent{
		MessageValueBytes: bytes,
	}, &Ziggurat{})

}
