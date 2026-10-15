// Package internalhttp is server.
package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// Server contains logger and calendar.
type Server struct {
	logger    Logger
	c         Calendar
	server    *http.Server
	mux       *http.ServeMux
	validator validator.Validate
}

// Logger is the logging interface used by HTTP server.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Calendar is for work with events.
type Calendar interface {
	CreateEvent(ctx context.Context, created event.Event) (event.Event, error)
	UpdateEvent(ctx context.Context, updated event.Event) (event.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	ListByDay(ctx context.Context, day time.Time) ([]event.Event, error)
	ListByWeek(ctx context.Context, week time.Time) ([]event.Event, error)
	ListByMonth(ctx context.Context, month time.Time) ([]event.Event, error)
}

// NewServer creates server.
func NewServer(logger Logger, app Calendar, host, port string) *Server {
	s := &Server{
		validator: *validator.New(),
		logger:    logger,
		c:         app,
		mux:       http.NewServeMux(),
	}
	// Register handlers and initialize the underlying server before publishing s.
	// Shutdown is then safe even if it wins the race with Start.
	s.mux.HandleFunc("GET /hello", loggingMiddleware(s.logger, hello))
	s.mux.HandleFunc("POST /event", loggingMiddleware(s.logger, s.createEvent))
	s.mux.HandleFunc("PUT /event", loggingMiddleware(s.logger, s.updateEvent))
	s.mux.HandleFunc("DELETE /event", loggingMiddleware(s.logger, s.deleteEvent))
	s.mux.HandleFunc("GET /event/day", loggingMiddleware(s.logger, s.getByDay))
	s.mux.HandleFunc("GET /event/week", loggingMiddleware(s.logger, s.getByWeek))
	s.mux.HandleFunc("GET /event/month", loggingMiddleware(s.logger, s.getByMonth))

	address := net.JoinHostPort(host, port)
	s.server = &http.Server{
		Addr:           address,
		Handler:        s.mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	return s
}

// Start serves requests until Stop is called or listening fails.
func (s *Server) Start(_ context.Context) error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		// Shutdown does not close active connections when its context expires.
		return errors.Join(err, s.server.Close())
	}
	return nil
}
