package event

import "github.com/google/uuid"

// FromCreateRequest maps CreateRequest to Event.
func FromCreateRequest(req CreateRequest) Event {
	event := Event{
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
func FromUpdateRequest(req UpdateRequest) Event {
	event := Event{}

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
func ToResponse(event Event) Response {
	return Response(event)
}

// ToResponses maps slice of Events to slice of Responses.
func ToResponses(events []Event) []Response {
	result := make([]Response, 0, len(events))

	for _, event := range events {
		result = append(result, ToResponse(event))
	}

	return result
}
