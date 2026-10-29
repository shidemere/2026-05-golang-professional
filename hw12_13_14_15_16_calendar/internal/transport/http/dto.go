package internalhttp

import (
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// CreateRequest is DTO for creating event.
// Description + NotifyBeforeSeconds is pointer type because they are optional.
// And pointer is easier for checking (instead of zero value in averate var).
type CreateRequest struct {
	Title               string    `json:"title" validate:"required,min=1,max=255"`
	ScheduledAt         time.Time `json:"scheduledAt" validate:"required"`
	FinishedAt          time.Time `json:"finishedAt" validate:"required,gtfield=ScheduledAt"`
	Description         *string   `json:"description" validate:"omitempty,max=2000"`
	UserID              uuid.UUID `json:"userId" validate:"required"`
	NotifyBeforeSeconds *int64    `json:"notifyBeforeSeconds" validate:"omitempty,gte=0"`
}

// UpdateRequest is DTO for updating event.
// All fields are pointers for nil checking.
type UpdateRequest struct {
	ID                  *uuid.UUID `json:"id" validate:"required"`
	Title               *string    `json:"title" validate:"required,min=1,max=255"`
	ScheduledAt         *time.Time `json:"scheduledAt" validate:"required"`
	FinishedAt          *time.Time `json:"finishedAt" validate:"required,gtfield=ScheduledAt"`
	Description         *string    `json:"description" validate:"omitempty,max=2000"`
	UserID              *uuid.UUID `json:"userId" validate:"required"`
	NotifyBeforeSeconds *int64     `json:"notifyBeforeSeconds" validate:"omitempty,gte=0"`
}

// Response is event DTO represent, not connected to database.
type Response struct {
	ID                  uuid.UUID `json:"id"`
	Title               string    `json:"title"`
	ScheduledAt         time.Time `json:"scheduledAt"`
	FinishedAt          time.Time `json:"finishedAt"`
	Description         string    `json:"description"`
	UserID              uuid.UUID `json:"userId"`
	NotifyBeforeSeconds int64     `json:"notifyBeforeSeconds"`
}

// FromCreateRequest maps CreateRequest to Event.
func FromCreateRequest(req CreateRequest) event.Event {
	event := event.Event{
		ID:          uuid.New(),
		Title:       req.Title,
		ScheduledAt: req.ScheduledAt,
		FinishedAt:  req.FinishedAt,
		UserID:      req.UserID,
	}

	if req.Description != nil {
		event.Description = *req.Description
	}

	if req.NotifyBeforeSeconds != nil {
		event.NotifyBeforeSeconds = *req.NotifyBeforeSeconds
	}

	return event
}

// FromUpdateRequest applies fields from UpdateRequest to existing Event.
func FromUpdateRequest(req UpdateRequest) event.Event {
	event := event.Event{}

	if req.ID != nil {
		event.ID = *req.ID
	}

	if req.Title != nil {
		event.Title = *req.Title
	}

	if req.ScheduledAt != nil {
		event.ScheduledAt = *req.ScheduledAt
	}

	if req.FinishedAt != nil {
		event.FinishedAt = *req.FinishedAt
	}

	if req.Description != nil {
		event.Description = *req.Description
	}

	if req.UserID != nil {
		event.UserID = *req.UserID
	}

	if req.NotifyBeforeSeconds != nil {
		event.NotifyBeforeSeconds = *req.NotifyBeforeSeconds
	}

	return event
}

// ToResponse maps Event to Response.
func ToResponse(event event.Event) Response {
	return Response{
		ID:                  event.ID,
		Title:               event.Title,
		ScheduledAt:         event.ScheduledAt,
		FinishedAt:          event.FinishedAt,
		Description:         event.Description,
		UserID:              event.ID,
		NotifyBeforeSeconds: event.NotifyBeforeSeconds,
	}
}

// ToResponses maps slice of Events to slice of Responses.
func ToResponses(events []event.Event) []Response {
	result := make([]Response, 0, len(events))

	for _, event := range events {
		result = append(result, ToResponse(event))
	}

	return result
}
