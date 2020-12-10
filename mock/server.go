package mock

import (
	"github.com/gojekfarm/ziggurat/ztype"
	"net/http"
)

type Server struct {
	ConfigureRoutesFunc func(a ztype.App, configFunc func(a ztype.App, h http.Handler))
	GetHandler          func() http.Handler
	StartFunc           func(a ztype.App) error
	StopFunc            func(a ztype.App)
}

func NewServer() *Server {
	return &Server{
		ConfigureRoutesFunc: func(a ztype.App, configFunc func(a ztype.App, h http.Handler)) {

		},
		GetHandler: func() http.Handler {
			return http.DefaultServeMux
		},
		StopFunc: func(a ztype.App) {

		},
		StartFunc: func(a ztype.App) error {
			return nil
		},
	}
}

func (s Server) ConfigureRoutes(a ztype.App, configFunc func(a ztype.App, h http.Handler)) {
	s.ConfigureRoutesFunc(a, configFunc)
}

func (s Server) Handler() http.Handler {
	return s.GetHandler()
}

func (s Server) Start(a ztype.App) error {
	return s.StartFunc(a)
}

func (s Server) Stop(a ztype.App) {
	s.StopFunc(a)
}
