package void

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"net/http"
)

type VoidServer struct{}

func NewVoidServer(c z.ConfigStore) z.HttpServer {
	return &VoidServer{}
}

func (v VoidServer) Start(app z.App) {
}

func (v VoidServer) ConfigureHTTPRoutes(a z.App, configFunc func(a z.App, h http.Handler)) {

}

func (v VoidServer) Stop(app z.App) error {
	return nil
}
