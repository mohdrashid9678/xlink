package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func New(handler http.Handler, cfg Config) *Server {
	readTimeout := cfg.ReadTimeout
	if readTimeout == 0 {
		readTimeout = defaultReadTimeout
	}

	writeTimeout := cfg.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = defaultWriteTimeout
	}

	idleTimeout := cfg.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = defaultIdleTimeout
	}

	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = defaultShutdownTimeout
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		shutdownTimeout: shutdownTimeout,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("http server listener error: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		log.Println("Shutdown signal received, draining active connections...")
		return s.gracefulShutdown()
	}
}

func (s *Server) gracefulShutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed, forced close: %w", err)
	}

	log.Println("HTTP server stopped gracefully")
	return nil
}
