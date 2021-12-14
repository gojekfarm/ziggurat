package server

import (
	"context"
	"net/http"

	"github.com/gojekfarm/ziggurat/logger"
	"github.com/julienschmidt/httprouter"
)

var defaultAddr = "localhost:8080"

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

func WithAddr(addr string) func(s *DefaultHttpServer) {
	return func(s *DefaultHttpServer) {
		s.server.Addr = addr
	}
}

// NewHTTPServer creates a new HTTPServer
// server.WithAddr can be used to configure the port number
// returns a ready to use http server with an embedded router
func NewHTTPServer(opts ...func(s *DefaultHttpServer)) *DefaultHttpServer {
	router := httprouter.New()
	requestLogger := logger.NOOP
	httpRequestLogger := HTTPRequestLogger(requestLogger)
	server := &http.Server{Handler: httpRequestLogger(router)}
	s := &DefaultHttpServer{
		server: server,
		router: router,
	}
	s.server.Addr = defaultAddr
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run method blocks until the server closes
// invoke it in a goroutine to run multiple servers concurrently
func (s *DefaultHttpServer) Run(ctx context.Context) error {
	errorChan := make(chan error)
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			errorChan <- err
		}
	}(s.server)

	go func() {
		done := ctx.Done()
		<-done
		if err := s.server.Shutdown(ctx); err != nil {
			errorChan <- err
		}
	}()

	return <-errorChan
}

// ConfigureHTTPEndpoints lets you configure your http endpoints
// a configuration function must be passed which will receive the
// router instance as an argument
func (s *DefaultHttpServer) ConfigureHTTPEndpoints(f func(r *httprouter.Router)) {
	f(s.router)
}

// ConfigureHandler lets you configure the embedded handler/router
// use this to add common middleware on the http router
func (s *DefaultHttpServer) ConfigureHandler(f func(r *httprouter.Router) http.Handler) {
	s.server.Handler = f(s.router)
}
