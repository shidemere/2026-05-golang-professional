package memorystorage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

func TestStorageCreateUpdateDelete(t *testing.T) {
	ctx := context.Background()
	storage := New()
	e := testEvent(time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC))

	created, err := storage.Create(ctx, e)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if created.ID != e.ID {
		t.Fatalf("created id = %s, want %s", created.ID, e.ID)
	}

	e.Title = "updated title"
	updated, err := storage.Update(ctx, e)
	if err != nil {
		t.Fatalf("update event: %v", err)
	}
	if updated.Title != e.Title {
		t.Fatalf("updated title = %q, want %q", updated.Title, e.Title)
	}

	if err := storage.Delete(ctx, e.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	if _, err := storage.Update(ctx, e); !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("update deleted event error = %v, want %v", err, event.ErrEventNotFound)
	}
}

func TestStorageListEvents(t *testing.T) {
	ctx := context.Background()
	storage := New()

	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	first := testEvent(day.Add(10 * time.Hour))
	second := testEvent(day.Add(15 * time.Hour))
	nextMonth := testEvent(day.AddDate(0, 1, 0))

	for _, e := range []event.Event{first, second, nextMonth} {
		if _, err := storage.Create(ctx, e); err != nil {
			t.Fatalf("create event %s: %v", e.ID, err)
		}
	}

	dayEvents, err := storage.GetByDay(ctx, day, day.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("get by day: %v", err)
	}
	assertEventIDs(t, dayEvents, first.ID, second.ID)

	weekEvents, err := storage.GetByWeek(ctx, day, day.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("get by week: %v", err)
	}
	assertEventIDs(t, weekEvents, first.ID, second.ID)

	monthEvents, err := storage.GetByMonth(ctx, day, day.AddDate(0, 1, 0))
	if err != nil {
		t.Fatalf("get by month: %v", err)
	}
	assertEventIDs(t, monthEvents, first.ID, second.ID)
}

func TestStorageBusinessErrors(t *testing.T) {
	ctx := context.Background()
	storage := New()

	base := testEvent(time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC))
	if _, err := storage.Create(ctx, base); err != nil {
		t.Fatalf("create base event: %v", err)
	}

	busy := testEvent(base.ScheduledAt.Add(30 * time.Minute))
	busy.UserID = base.UserID
	if _, err := storage.Create(ctx, busy); !errors.Is(err, event.ErrDateBusy) {
		t.Fatalf("create busy event error = %v, want %v", err, event.ErrDateBusy)
	}

	other := testEvent(base.ScheduledAt.Add(3 * time.Hour))
	if _, err := storage.Create(ctx, other); err != nil {
		t.Fatalf("create other event: %v", err)
	}
	other.ScheduledAt = base.ScheduledAt.Add(30 * time.Minute)
	other.FinishedAt = base.FinishedAt.Add(30 * time.Minute)
	other.UserID = base.UserID
	if _, err := storage.Update(ctx, other); !errors.Is(err, event.ErrDateBusy) {
		t.Fatalf("update busy event error = %v, want %v", err, event.ErrDateBusy)
	}

	if err := storage.Delete(ctx, uuid.New()); !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("delete unknown event error = %v, want %v", err, event.ErrEventNotFound)
	}
}

func TestStorageConcurrentSafety(t *testing.T) {
	ctx := context.Background()
	storage := New()
	start := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			e := testEvent(start.Add(time.Duration(i) * time.Hour))
			if _, err := storage.Create(ctx, e); err != nil {
				t.Errorf("create event %d: %v", i, err)
				return
			}

			e.Title = "updated"
			if _, err := storage.Update(ctx, e); err != nil {
				t.Errorf("update event %d: %v", i, err)
				return
			}

			if _, err := storage.GetByDay(ctx, start, start.AddDate(0, 0, 7)); err != nil {
				t.Errorf("get by day %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
}

func testEvent(start time.Time) event.Event {
	return event.Event{
		ID:                  uuid.New(),
		Title:               "event",
		ScheduledAt:         start,
		FinishedAt:          start.Add(time.Hour),
		Description:         "description",
		UserID:              uuid.New(),
		NotifyBeforeSeconds: int64(15 * time.Minute / time.Second),
	}
}

func assertEventIDs(t *testing.T, events []event.Event, want ...uuid.UUID) {
	t.Helper()

	got := make(map[uuid.UUID]struct{}, len(events))
	for _, e := range events {
		got[e.ID] = struct{}{}
	}

	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d", len(got), len(want))
	}
	for _, id := range want {
		if _, ok := got[id]; !ok {
			t.Fatalf("event %s not found in result", id)
		}
	}
}
