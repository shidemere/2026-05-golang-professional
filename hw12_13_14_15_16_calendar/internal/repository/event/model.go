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
	Status              Status
}

// Status reppresents lifecycle for scheduler and sender.
// Can be:
// PENDING -> default value for new Event.
// PROCESSING -> scheduler take it from database for processing.
// COMPLETED -> sender processed it.
// FAILED -> some error occurred.
type Status struct {
	Status string
}

func NewPending() Status {
	return Status{Status: "PENDING"}
}
