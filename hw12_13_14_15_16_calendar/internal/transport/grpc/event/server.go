package internalgrpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server handles calendar RPCs.
type Server struct {
	generated.UnimplementedEventServiceServer
	logger Logger
	c      Calendar
}

// NewServer creates a calendar gRPC server.
func NewServer(logger Logger, calendar Calendar) *Server {
	return &Server{
		logger: logger,
		c:      calendar,
	}
}

// Logger is the logging interface used by the gRPC server.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Calendar defines the application operations used by the server.
type Calendar interface {
	CreateEvent(ctx context.Context, created event.Event) (event.Event, error)
	UpdateEvent(ctx context.Context, updated event.Event) (event.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	ListByDay(ctx context.Context, day time.Time) ([]event.Event, error)
	ListByWeek(ctx context.Context, week time.Time) ([]event.Event, error)
	ListByMonth(ctx context.Context, month time.Time) ([]event.Event, error)
}

// CreateEvent creates an event.
func (s *Server) CreateEvent(ctx context.Context, request *generated.CreateEventRequst) (*generated.Event, error) {
	converted, err := fromCreateToEvent(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.c.CreateEvent(ctx, converted)
	if err != nil {
		return nil, s.serviceError(err)
	}

	response := fromModelToEvent(result)
	return &response, nil
}

// UpdateEvent updates an event.
func (s *Server) UpdateEvent(ctx context.Context, request *generated.UpdateEventRequest) (*generated.Event, error) {
	converted, err := fromUpdateToEvent(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result, err := s.c.UpdateEvent(ctx, converted)
	if err != nil {
		return nil, s.serviceError(err)
	}
	response := fromModelToEvent(result)
	return &response, nil
}

// DeleteEvent deletes an event.
func (s *Server) DeleteEvent(ctx context.Context, request *generated.DeleteEventRequest) (*emptypb.Empty, error) {
	if request == nil || request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id for deleting can't be empty")
	}

	if _, err := uuid.Parse(request.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event id")
	}
	err := s.c.DeleteEvent(ctx, request.Id)
	if err != nil {
		return nil, s.serviceError(err)
	}
	return &emptypb.Empty{}, nil
}

// ListByDay lists events for a day.
func (s *Server) ListByDay(ctx context.Context, day *generated.Day) (*generated.Events, error) {
	if day == nil || day.Day == nil {
		return nil, status.Error(codes.InvalidArgument, "a valid timestamp is required")
	}
	if err := day.Day.CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid day: %v", err)
	}
	time := day.Day.AsTime()
	events, err := s.c.ListByDay(ctx, time)
	if err != nil {
		return nil, s.serviceError(err)
	}
	response := fromModelsToEvents(events)
	return response, nil
}

// ListByWeek lists events for a week.
func (s *Server) ListByWeek(ctx context.Context, week *generated.Week) (*generated.Events, error) {
	if week == nil || week.Week == nil {
		return nil, status.Error(codes.InvalidArgument, "a valid timestamp is required")
	}
	if err := week.Week.CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid week: %v", err)
	}
	time := week.Week.AsTime()
	events, err := s.c.ListByWeek(ctx, time)
	if err != nil {
		return nil, s.serviceError(err)
	}
	response := fromModelsToEvents(events)
	return response, nil
}

// ListByMonth lists events for a month.
func (s *Server) ListByMonth(ctx context.Context, month *generated.Month) (*generated.Events, error) {
	if month == nil || month.Month == nil {
		return nil, status.Error(codes.InvalidArgument, "a valid timestamp is required")
	}
	if err := month.Month.CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid month: %v", err)
	}
	time := month.Month.AsTime()
	events, err := s.c.ListByMonth(ctx, time)
	if err != nil {
		return nil, s.serviceError(err)
	}
	response := fromModelsToEvents(events)
	return response, nil
}

// serviceError translates application errors at the transport boundary.
func (s *Server) serviceError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, event.ErrEventNotFound):
		return status.Error(codes.NotFound, "event not found")
	case errors.Is(err, event.ErrDateBusy):
		return status.Error(codes.AlreadyExists, "event time is occupied")
	default:
		s.logger.Warn("calendar request failed", "err", err)
		return status.Error(codes.Internal, "internal server error")
	}
}
