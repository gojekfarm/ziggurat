package server

import (
	"context"
	"net/http"
)

func Run(ctx context.Context, s *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
	select {
	case <-ctx.Done():
		return s.Shutdown(ctx)
	case err := <-errChan:
		return err
	}
}
