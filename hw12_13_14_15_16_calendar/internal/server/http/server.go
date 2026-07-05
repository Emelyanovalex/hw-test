package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
)

// Logger is the interface used by the HTTP server for logging.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Application is the business-logic interface used by HTTP handlers.
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

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
	h := &handler{app: app, logger: logger}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /events", h.createEvent)
	mux.HandleFunc("PUT /events/{id}", h.updateEvent)
	mux.HandleFunc("DELETE /events/{id}", h.deleteEvent)
	mux.HandleFunc("GET /events/day", h.listEventsForDay)
	mux.HandleFunc("GET /events/week", h.listEventsForWeek)
	mux.HandleFunc("GET /events/month", h.listEventsForMonth)

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	srv := &http.Server{
		Addr:         addr,
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
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
