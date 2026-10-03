// Package internalhttp is server.
package internalhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// Server contains logger and calendar.
type Server struct { // TODO
	server *http.Server
	Logger
	c Calendar
}

type Logger interface { // TODO
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
func NewServer(logger Logger, app Calendar) *Server {
	return &Server{}
}

// Start server.
func (s *Server) Start(ctx context.Context) error {
	// TODO
	http.HandleFunc("/hello", nil) // TODO: fill

	err := s.server.ListenAndServe()
	<-ctx.Done()
	return err
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// TODO
