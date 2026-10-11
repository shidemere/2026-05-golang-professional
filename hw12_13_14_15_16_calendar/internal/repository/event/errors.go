package event

import "errors"

var (
	// ErrDateBusy means that another event already occupies this time interval.
	ErrDateBusy = errors.New("event date is busy")
	// ErrEventNotFound means that event does not exist in storage.
	ErrEventNotFound = errors.New("event not found")
)
