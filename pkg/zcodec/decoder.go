package zcodec

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
)

func JSON(bytes []byte, model interface{}) error {
	return json.Unmarshal(bytes, model)
}

func Proto(bytes []byte, model proto.Message) error {
	return proto.Unmarshal(bytes, model)
}
