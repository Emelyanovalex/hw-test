package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Logger is the interface used by the HTTP server for logging.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Application is intentionally narrow at this stage — HW12 only requires a
// hello-world handler that is independent of the business logic. We keep the
// interface so subsequent homework can add real handlers.
type Application interface{}

type Server struct {
	httpServer *http.Server
	logger     Logger
}

// Config holds HTTP-server-specific configuration.
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func NewServer(logger Logger, app Application, cfg Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/", helloHandler)

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	srv := &http.Server{
		Addr:         addr,
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	_ = app // placeholder until business handlers land
	return &Server{httpServer: srv, logger: logger}
}

func (s *Server) Start(_ context.Context) error {
	s.logger.Info(fmt.Sprintf("http server listening on %s", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server is shutting down")
	return s.httpServer.Shutdown(ctx)
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello-world"))
}
