// Package app is service layer.
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// App is main service layer.
type App struct {
	log     Logger
	storage Storage
}

// Logger containing only necessaery functions.
type Logger interface {
	Warn(msg string, args ...any)
}

// Storage contains all storages.
type Storage interface {
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
		log:     logger,
		storage: storage,
	}
}

// CreateEvent by request withoud id.
func (a *App) CreateEvent(ctx context.Context, created event.Event) (event.Event, error) {
	created.ID = uuid.New()
	e, err := a.storage.Create(ctx, created)
	if err != nil {
		a.log.Warn("error occurred in storage while creating", "err", err)
		return event.Event{}, err
	}
	return e, nil
}

// UpdateEvent event from request by id.
func (a *App) UpdateEvent(ctx context.Context, updated event.Event) (event.Event, error) {
	u, err := a.storage.Update(ctx, updated)
	if err != nil {
		a.log.Warn("can't update event", "err", err)
		return event.Event{}, err
	}
	return u, nil
}

// DeleteEvent delete event by id.
func (a *App) DeleteEvent(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		a.log.Warn("incorrect uuid", "id", id)
		return err
	}
	err = a.storage.Delete(ctx, uuid)
	if err != nil {
		a.log.Warn("can't delete event by id", "err", err)
		return err
	}

	return nil
}

// ListByDay get all events by day.
func (a *App) ListByDay(ctx context.Context, day time.Time) ([]event.Event, error) {
	startOfDay := time.Date(
		day.Year(),
		day.Month(),
		day.Day(),
		0, 0, 0, 0,
		day.Location(),
	)

	startOfNextDay := startOfDay.AddDate(0, 0, 1)

	log.Info("startOfDay is :", startOfDay)
	log.Info("startOfNextDay is :", startOfNextDay)
	events, err := a.storage.GetByDay(ctx, startOfDay, startOfNextDay)
	if err != nil {
		a.log.Warn("can't get events by day", "err", err)
		return nil, err
	}
	return events, nil
}

// ListByWeek get all events by week.
func (a *App) ListByWeek(ctx context.Context, week time.Time) ([]event.Event, error) {
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

	events, err := a.storage.GetByWeek(ctx, startOfWeek, startOfNextWeek)
	if err != nil {
		a.log.Warn("can't get events by week", "err", err)

		return nil, err
	}

	return events, nil
}

// ListByMonth get all events by month.
func (a *App) ListByMonth(ctx context.Context, month time.Time) ([]event.Event, error) {
	startOfMonth := time.Date(
		month.Year(),
		month.Month(),
		1,
		0, 0, 0, 0,
		month.Location(),
	)

	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	events, err := a.storage.GetByMonth(ctx, startOfMonth, startOfNextMonth)
	if err != nil {
		a.log.Warn("can't get events by month", "err", err)

		return nil, err
	}

	return events, nil
}
