// Package internalgrpc implements the calendar gRPC transport.
package internalgrpc

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event/generated"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func fromCreateToEvent(in *generated.CreateEventRequst) (event.Event, error) {
	if in == nil {
		return event.Event{}, fmt.Errorf("event request is required")
	}
	if in.Title == "" || in.FinishedAt == nil || in.ScheduledAt == nil || in.UserId == "" {
		return event.Event{}, fmt.Errorf("validation error: title, finished_at, schedulet_at, user_id is necessary")
	}
	if err := validateEventTimes(in.ScheduledAt, in.FinishedAt, in.NotifyBeforeSeconds); err != nil {
		return event.Event{}, err
	}

	userID, err := uuid.Parse(in.UserId)
	if err != nil {
		return event.Event{}, fmt.Errorf("can't parse user_id: %w", err)
	}

	return event.Event{
		Title:               in.Title,
		FinishedAt:          in.FinishedAt.AsTime(),
		ScheduledAt:         in.ScheduledAt.AsTime(),
		UserID:              userID,
		Description:         in.Description,
		NotifyBeforeSeconds: in.NotifyBeforeSeconds,
	}, nil
}

func fromUpdateToEvent(in *generated.UpdateEventRequest) (event.Event, error) {
	if in == nil {
		return event.Event{}, fmt.Errorf("event request is required")
	}
	if in.Id == "" || in.Title == "" || in.FinishedAt == nil || in.ScheduledAt == nil || in.UserId == "" {
		return event.Event{}, fmt.Errorf("validation error: title, finished_at, schedulet_at, user_id is necessary")
	}
	if err := validateEventTimes(in.ScheduledAt, in.FinishedAt, in.NotifyBeforeSeconds); err != nil {
		return event.Event{}, err
	}

	id, err := uuid.Parse(in.Id)
	if err != nil {
		return event.Event{}, fmt.Errorf("can't parse id: %w", err)
	}

	userID, err := uuid.Parse(in.UserId)
	if err != nil {
		return event.Event{}, fmt.Errorf("can't parse user_id: %w", err)
	}

	return event.Event{
		ID:                  id,
		Title:               in.Title,
		FinishedAt:          in.FinishedAt.AsTime(),
		ScheduledAt:         in.ScheduledAt.AsTime(),
		UserID:              userID,
		Description:         in.Description,
		NotifyBeforeSeconds: in.NotifyBeforeSeconds,
	}, nil
}

func validateEventTimes(start, end *timestamppb.Timestamp, notifyBefore int64) error {
	if err := start.CheckValid(); err != nil {
		return fmt.Errorf("invalid scheduled_at: %w", err)
	}
	if err := end.CheckValid(); err != nil {
		return fmt.Errorf("invalid finished_at: %w", err)
	}
	if !end.AsTime().After(start.AsTime()) {
		return fmt.Errorf("finished_at must be after scheduled_at")
	}
	if notifyBefore < 0 {
		return fmt.Errorf("notify_before_seconds must not be negative")
	}
	return nil
}

func fromModelToEvent(in event.Event) generated.Event {
	return generated.Event{
		Id:                  in.ID.String(),
		Title:               in.Title,
		FinishedAt:          timestamppb.New(in.FinishedAt),
		ScheduledAt:         timestamppb.New(in.ScheduledAt),
		UserId:              in.UserID.String(),
		Description:         in.Description,
		NotifyBeforeSeconds: in.NotifyBeforeSeconds,
	}
}

func fromModelsToEvents(events []event.Event) *generated.Events {
	result := &generated.Events{}
	result.Event = make([]*generated.Event, 0, len(events))
	for _, v := range events {
		response := fromModelToEvent(v)
		result.Event = append(result.Event, &response)
	}
	return result
}
