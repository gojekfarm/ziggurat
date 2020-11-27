package void

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"net/http"
)

type VoidServer struct{}

func NewServer(c z.ConfigStore) z.Server {
	return &VoidServer{}
}

func (v VoidServer) Start(app z.App) error {
	return nil
}

func (v VoidServer) ConfigureRoutes(a z.App, configFunc func(a z.App, h http.Handler)) {

}

func (v VoidServer) Stop(app z.App) {

}

func (v VoidServer) Handler() http.Handler {
	return nil
}
