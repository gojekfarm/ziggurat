package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	hs := &http.Server{Addr: ":8080"}
	errCh := make(chan error)
	ctx, cfn := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cfn()
	go func() {
		err := Run(ctx, hs)
		errCh <- err
	}()
	_, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Errorf("error starting http server on 8080: %v", err)
	}
	if err := <-errCh; !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("expected error [%v] got [%v]", context.DeadlineExceeded, err)
	}
}
