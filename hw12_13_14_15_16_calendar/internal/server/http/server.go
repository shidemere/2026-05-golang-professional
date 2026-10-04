// Package internalhttp is server.
package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// Server contains logger and calendar.
type Server struct { // TODO
	host   string
	port   string
	logger Logger
	c      Calendar
	s      *echo.Echo
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
		s:      echo.New(),
		host:   host,
		port:   port,
	}
}

// Start server.
func (s *Server) Start(ctx context.Context) error {
	s.s.GET("/hello", hello, loggingMiddleware(s.logger))
	// http.Handle("/hello", loggingMiddleware(http.HandlerFunc(hello)))
	address := fmt.Sprintf("%s:%s", s.host, s.port)
	err := s.s.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.s.Logger.Error("some shit happen with server", err)
		log.Fatalf("closing server")
	}
	<-ctx.Done()
	return err
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error {
	err := s.s.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
