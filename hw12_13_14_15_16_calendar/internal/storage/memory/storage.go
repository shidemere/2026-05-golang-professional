// Package memorystorage is for containing events in memory.
package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

// Storage is in-memory storage with map insise. Conccurent safe.
type Storage struct {
	cache map[uuid.UUID]event.Event
	mu    sync.RWMutex
}

// New create a new in-memory storage.
func New() *Storage {
	return &Storage{
		cache: make(map[uuid.UUID]event.Event),
	}
}

// Create a new item.
func (s *Storage) Create(_ context.Context, e event.Event) (event.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[e.ID] = e
	return e, nil
}

// Update item.
// TODO: is any way don't get twice?
func (s *Storage) Update(_ context.Context, e event.Event) (event.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.cache[e.ID]
	if !ok {
		return event.Event{}, fmt.Errorf("not found event for update")
	}
	s.cache[e.ID] = e
	return e, nil
}

// Delete item.
func (s *Storage) Delete(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, id)
	return nil
}

// GetByDay provides all events scheduled at certain day.
func (s *Storage) GetByDay(_ context.Context, startOfDay, startOfNextDay time.Time) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]event.Event, 0)

	for _, v := range s.cache {
		if !v.ScheduledAt.Before(startOfDay) &&
			v.ScheduledAt.Before(startOfNextDay) {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no events scheduled at this day")
	}
	return result, nil
}

// GetByWeek provides all events scheduled at certain week.
func (s *Storage) GetByWeek(_ context.Context, startOfWeek, endOfWeek time.Time) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]event.Event, 0)

	for _, v := range s.cache {
		if !v.ScheduledAt.Before(startOfWeek) &&
			v.ScheduledAt.Before(endOfWeek) {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no events scheduled at this week")
	}
	return result, nil
}

// GetByMonth provides all events scheduled at certain month.
func (s *Storage) GetByMonth(_ context.Context, startOfMonth, endOfMonth time.Time) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]event.Event, 0)

	for _, v := range s.cache {
		if !v.ScheduledAt.Before(startOfMonth) &&
			v.ScheduledAt.Before(endOfMonth) {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no events scheduled at this month")
	}
	return result, nil
}
