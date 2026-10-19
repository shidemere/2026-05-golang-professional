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
