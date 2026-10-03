// Package app is service layer.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// App is main service layer.
type App struct { // TODO
	validator validator.Validate
	log       Logger
	storage   Storage
}

// Logger containing only necessaery functions.
type Logger interface { // TODO
	Warn(msg string, args ...any)
}

// Storage contains all storages.
type Storage interface { // TODO
	EventStorage
}

// EventStorage is abstraction above events storage.
type EventStorage interface {
	Create(ctx context.Context, event event.Event) (event.Event, error)
	Update(ctx context.Context, event event.Event) (event.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByDay(ctx context.Context, startOfDay, startOfNextDay time.Time) ([]event.Event, error)
	GetByWeek(ctx context.Context, startOfWeek, endOfWeek time.Time) ([]event.Event, error)
	GetByMonth(ctx context.Context, startOfMonth, endOfMonth time.Time) ([]event.Event, error)
}

// New is constructor for App.
func New(logger Logger, storage Storage) *App {
	return &App{
		validator: *validator.New(),
		log:       logger,
		storage:   storage,
	}
}

// CreateEvent by request withoud id.
func (a *App) CreateEvent(ctx context.Context, created event.CreateRequest) event.WrapperResponse {
	if err := a.validator.Struct(created); err != nil {
		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			a.log.Warn("validation error occurred", "validation error", err)
			return event.WrapperResponse{
				Err:  err.Error(),
				Resp: event.Response{},
			}
		}
		a.log.Warn("some shit happen", "err", err)
		return event.WrapperResponse{
			Err:  err.Error(),
			Resp: event.Response{},
		}
	}

	e := event.FromCreateRequest(created)
	e.ID = uuid.New()
	e, err := a.storage.Create(ctx, e)
	if err != nil {
		a.log.Warn("some shit happen", "err", err)
		return event.WrapperResponse{
			Err:  err.Error(),
			Resp: event.Response{},
		}
	}
	response := event.ToResponse(e)
	return event.WrapperResponse{Err: "", Resp: response}
}

// UpdateEvent event from request by id.
func (a *App) UpdateEvent(ctx context.Context, updated event.UpdateRequest) event.WrapperResponse {
	if err := a.validator.Struct(updated); err != nil {
		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			a.log.Warn("validation error occurred", "validation error", err)
			return event.WrapperResponse{Err: err.Error(), Resp: event.Response{}}
		}
		a.log.Warn("some shit happen", "err", err)
		return event.WrapperResponse{
			Err:  err.Error(),
			Resp: event.Response{},
		}
	}

	u := event.FromUpdateRequest(updated)
	u, err := a.storage.Update(ctx, u)
	if err != nil {
		a.log.Warn("some shit happen", "err", err)
		return event.WrapperResponse{
			Err:  err.Error(),
			Resp: event.Response{},
		}
	}
	response := event.ToResponse(u)
	return event.WrapperResponse{Err: "", Resp: response}
}

// DeleteEvent delete event by id.
func (a *App) DeleteEvent(ctx context.Context, id string) event.WrapperResponse {
	uuid, err := uuid.Parse(id)
	if err != nil {
		a.log.Warn("incorrect uuid", "id", id)
		return event.WrapperResponse{Err: err.Error(), Resp: event.Response{}}
	}
	err = a.storage.Delete(ctx, uuid)
	if err != nil {
		a.log.Warn("can't delete event by id", "err", err)
		return event.WrapperResponse{Err: err.Error(), Resp: event.Response{}}
	}

	return event.WrapperResponse{Err: "", Resp: event.Response{}}
}

// ListByDay get all events by day.
func (a *App) ListByDay(ctx context.Context, day time.Time) event.WrapperResponses {
	startOfDay := time.Date(
		day.Year(),
		day.Month(),
		day.Day(),
		0, 0, 0, 0,
		day.Location(),
	)

	startOfNextDay := startOfDay.AddDate(0, 0, 1)

	events, err := a.storage.GetByDay(ctx, startOfDay, startOfNextDay)
	if err != nil {
		a.log.Warn("can't get events by day", "err", err)
	}
	responses := event.ToResponses(events)
	return event.WrapperResponses{Err: "", Resp: responses}
}

// ListByWeek get all events by week.
func (a *App) ListByWeek(ctx context.Context, week time.Time) event.WrapperResponses {
	weekday := int(week.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	startOfWeek := time.Date(
		week.Year(),
		week.Month(),
		week.Day()-weekday+1,
		0, 0, 0, 0,
		week.Location(),
	)

	startOfNextWeek := startOfWeek.AddDate(0, 0, 7)

	events, err := a.storage.GetByDay(ctx, startOfWeek, startOfNextWeek)
	if err != nil {
		a.log.Warn("can't get events by week", "err", err)

		return event.WrapperResponses{
			Err:  err.Error(),
			Resp: nil,
		}
	}

	responses := event.ToResponses(events)

	return event.WrapperResponses{
		Err:  "",
		Resp: responses,
	}
}

// ListByMonth get all events by month.
func (a *App) ListByMonth(ctx context.Context, month time.Time) event.WrapperResponses {
	startOfMonth := time.Date(
		month.Year(),
		month.Month(),
		1,
		0, 0, 0, 0,
		month.Location(),
	)

	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	events, err := a.storage.GetByDay(ctx, startOfMonth, startOfNextMonth)
	if err != nil {
		a.log.Warn("can't get events by month", "err", err)

		return event.WrapperResponses{
			Err:  err.Error(),
			Resp: nil,
		}
	}

	responses := event.ToResponses(events)

	return event.WrapperResponses{
		Err:  "",
		Resp: responses,
	}
}
