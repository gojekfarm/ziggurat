package prometheus

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startMonitoringServer(ctx context.Context, h http.Handler, opts ...ServerOpts) error {
	ctxWithCancel, cfn := context.WithCancel(ctx)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	defer cfn()
	server := &http.Server{
		Addr:    "localhost:9090",
		Handler: h,
	}

	for _, o := range opts {
		o(server)
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
