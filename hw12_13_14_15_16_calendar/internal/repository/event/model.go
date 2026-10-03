// Package event contains all for work with events from calendar.
package event

import (
	"time"

	"github.com/google/uuid"
)

// Event reppresent some action in calendar.
type Event struct {
	ID                  uuid.UUID `db:"id"`
	Title               string    `db:"title"`
	ScheduledAt         time.Time `db:"scheduled_at"`
	FinishedAt          time.Time `db:"finished_at"`
	Description         string    `db:"description"`
	UserID              uuid.UUID `db:"user_id"`
	NotifyBeforeSeconds int64     `db:"notify_before_seconds"`
}

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
	Title               *string    `json:"title" validate:"omitempty,min=1,max=255"`
	ScheduledAt         *time.Time `json:"scheduledAt" validate:"omitempty"`
	FinishedAt          *time.Time `json:"finishedAt" validate:"omitempty"`
	Description         *string    `json:"description" validate:"omitempty,max=2000"`
	UserID              *uuid.UUID `json:"userId" validate:"omitempty"`
	NotifyBeforeSeconds *int64     `json:"notifyBeforeSeconds" validate:"omitempty,gte=0"`
}

// WrapperResponse is wrapper around Response with error.
type WrapperResponse struct {
	Err  string   `json:"error,omitempty"`
	Resp Response `json:"response,omitempty"`
}

// WrapperResponses is wrapper around Response with error.
type WrapperResponses struct {
	Err  string     `json:"error,omitempty"`
	Resp []Response `json:"response,omitempty"`
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
