package mock

import (
	"github.com/gojekfarm/ziggurat/z"
	"net/http"
)

type Server struct {
	ConfigureRoutesFunc func(a z.App, configFunc func(a z.App, h http.Handler))
	GetHandler          func() http.Handler
	StartFunc           func(a z.App) error
	StopFunc            func(a z.App)
}

func NewServer() *Server {
	return &Server{
		ConfigureRoutesFunc: func(a z.App, configFunc func(a z.App, h http.Handler)) {

		},
		GetHandler: func() http.Handler {
			return http.DefaultServeMux
		},
		StopFunc: func(a z.App) {

		},
		StartFunc: func(a z.App) error {
			return nil
		},
	}
}

func (s Server) ConfigureRoutes(a z.App, configFunc func(a z.App, h http.Handler)) {
	s.ConfigureRoutesFunc(a, configFunc)
}

func (s Server) Handler() http.Handler {
	return s.GetHandler()
}

func (s Server) Start(a z.App) error {
	return s.StartFunc(a)
}

func (s Server) Stop(a z.App) {
	s.StopFunc(a)
}
