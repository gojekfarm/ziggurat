package mw

import (
	"encoding/json"
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/golang/protobuf/proto"
)

func ProtoDecoder(next z.HandlerFunc) z.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		messageEvent.ValueDecoderHook = func(model interface{}) error {
			if protoMessage, ok := model.(proto.Message); !ok {
				return errors.New("model not of type `proto.Message`")
			} else {
				return proto.Unmarshal(messageEvent.MessageValueBytes, protoMessage)
			}
		}
		return next(messageEvent, app)
	}
}

func JSONDecoder(next z.HandlerFunc) z.HandlerFunc {
	return func(message basic.MessageEvent, app z.App) z.ProcessStatus {
		message.ValueDecoderHook = func(model interface{}) error {
			return json.Unmarshal(message.MessageValueBytes, model)
		}
		return next(message, app)
	}
}
