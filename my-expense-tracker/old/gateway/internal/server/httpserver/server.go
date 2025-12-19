package httpserver

import (
	"context"
	"errors"
	nethttp "net/http"
	"time"

	"gitlab.com/education/gateway/internal/config"
)

type Server struct {
	srv             *nethttp.Server
	shutdownTimeout time.Duration
}

func New(cfg config.HTTPConfig, handler nethttp.Handler) *Server {
	return &Server{
		srv: &nethttp.Server{
			Addr:         cfg.Address,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *Server) Start() error {
	if s == nil || s.srv == nil {
		return errors.New("http server is not initialized")
	}
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	timeout := s.shutdownTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
