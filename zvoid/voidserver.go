package zvoid

import (
	"github.com/gojekfarm/ziggurat/ztype"
	"net/http"
)

type VoidServer struct{}

func NewServer() ztype.Server {
	return &VoidServer{}
}

func (v VoidServer) Start(app ztype.App) error {
	return nil
}

func (v VoidServer) ConfigureRoutes(a ztype.App, configFunc func(a ztype.App, h http.Handler)) {

}

func (v VoidServer) Stop(app ztype.App) {

}

func (v VoidServer) Handler() http.Handler {
	return nil
}
