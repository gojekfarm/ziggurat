package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func startMonitoringServer(ctx context.Context, port string, h http.Handler) error {
	ctxWithCancel, cfn := context.WithCancel(ctx)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	defer cfn()
	server := &http.Server{
		Addr:    "localhost" + port,
		Handler: h,
	}
	errCh := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		return server.Shutdown(ctxWithCancel)
	case err := <-errCh:
		return err
	}
}
