// Package internalhttp is server.
package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// Server contains logger and calendar.
type Server struct {
	host   string
	port   string
	logger Logger
	c      Calendar
	server *http.Server
	mux    *http.ServeMux
}

// Logger is the logging interface used by HTTP server.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Calendar is for work with events.
type Calendar interface {
	CreateEvent(ctx context.Context, created event.CreateRequest) event.WrapperResponse
	UpdateEvent(ctx context.Context, updated event.UpdateRequest) event.WrapperResponse
	DeleteEvent(ctx context.Context, id string) event.WrapperResponse
	ListByDay(ctx context.Context, day time.Time) event.WrapperResponses
	ListByWeek(ctx context.Context, week time.Time) event.WrapperResponses
	ListByMonth(ctx context.Context, month time.Time) event.WrapperResponses
}

// NewServer creates server.
func NewServer(logger Logger, app Calendar, host, port string) *Server {
	return &Server{
		logger: logger,
		c:      app,
		mux:    http.NewServeMux(),
		host:   host,
		port:   port,
	}
}

// Start server.
func (s *Server) Start(ctx context.Context) error {
	// handler registration
	s.mux.HandleFunc("GET /hello", loggingMiddleware(s.logger, hello))
	s.mux.HandleFunc("POST /event", s.createEvent)
	s.mux.HandleFunc("PUT /event", s.updateEvent)
	s.mux.HandleFunc("DELETE /event", s.deleteEvent)
	s.mux.HandleFunc("GET /event/day", s.getByDay)
	s.mux.HandleFunc("GET /event/week", s.getByWeek)
	s.mux.HandleFunc("GET /event/month", s.getByMonth)

	address := fmt.Sprintf("%s:%s", s.host, s.port)
	s.server = &http.Server{
		Addr:           address,
		Handler:        s.mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Warn("error occurred in web-server", err)
		log.Fatalf("closing server")
	}
	<-ctx.Done()
	return err
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
