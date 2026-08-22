package server

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestServerGracefulShutdownOnContextCancel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := New(mux, Config{
		Port:            "0", // Dynamic available port
		ShutdownTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Trigger graceful shutdown by canceling context
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error on graceful shutdown, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server failed to shutdown within timeout")
	}
}
