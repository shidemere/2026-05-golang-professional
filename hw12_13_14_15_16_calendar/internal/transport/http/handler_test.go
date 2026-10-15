package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/app"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	memory "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/storage/memory"
)

const invalidValue = "wrong"

func testServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(logger, app.New(logger, memory.New()), "", "")
}

func eventBody() map[string]any {
	return map[string]any{
		"title": "Meeting", "scheduledAt": "2026-10-18T10:00:00Z", "finishedAt": "2026-10-18T11:00:00Z",
		"userId": "22222222-2222-4222-8222-222222222222", "description": "Agenda", "notifyBeforeSeconds": 60,
	}
}

func call(t *testing.T, handler http.HandlerFunc, method, target string, body any, want int) []byte {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequestWithContext(t.Context(), method, target, bytes.NewReader(data)))
	if w.Code != want {
		t.Fatalf("%s %s: status %d, want %d; body: %s", method, target, w.Code, want, w.Body.String())
	}
	return w.Body.Bytes()
}

func TestHTTPLifecycle(t *testing.T) {
	s := testServer()
	body := eventBody()
	data := call(t, s.createEvent, "POST", "/event", body, 200)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"id", "title", "scheduledAt", "finishedAt", "userId", "description", "notifyBeforeSeconds",
	} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("missing JSON field %s: %s", key, data)
		}
	}
	if _, ok := raw["ID"]; ok {
		t.Fatal("domain model leaked into JSON response")
	}
	var created Response
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("empty created ID")
	}
	body["id"] = created.ID.String()
	body["title"] = "Updated"
	body["scheduledAt"] = "2026-10-18T12:00:00Z"
	body["finishedAt"] = "2026-10-18T13:00:00Z"
	body["userId"] = uuid.NewString()
	body["description"] = "Changed"
	body["notifyBeforeSeconds"] = 120
	updatedData := call(t, s.updateEvent, "PUT", "/event", body, 200)
	var updated Response
	if err := json.Unmarshal(updatedData, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.ID != created.ID ||
		updated.Title != "Updated" ||
		updated.UserID.String() != body["userId"] ||
		updated.Description != "Changed" ||
		updated.NotifyBeforeSeconds != 120 ||
		updated.ScheduledAt.Format(time.RFC3339) != body["scheduledAt"] ||
		updated.FinishedAt.Format(time.RFC3339) != body["finishedAt"] {
		t.Fatalf("update lost fields: %+v", updated)
	}
	for _, handler := range []http.HandlerFunc{s.getByDay, s.getByWeek, s.getByMonth} {
		data = call(t, handler, "GET", "/event?date=2026-10-18", nil, 200)
		var events []Response
		if err := json.Unmarshal(data, &events); err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 || events[0] != updated {
			t.Fatalf("saved event differs: %+v", events)
		}
	}
	call(t, s.deleteEvent, "DELETE", "/event?id="+created.ID.String(), nil, 202)
	call(t, s.getByDay, "GET", "/event/day?date=2026-10-18", nil, 204)
	call(t, s.deleteEvent, "DELETE", "/event?id="+created.ID.String(), nil, 404)
	call(t, s.updateEvent, "PUT", "/event", body, 404)
}

func TestHTTPValidation(t *testing.T) {
	cases := []struct {
		name   string
		change func(map[string]any)
	}{
		{"missing title", func(b map[string]any) { delete(b, "title") }},
		{"empty title", func(b map[string]any) { b["title"] = "" }},
		{"missing start", func(b map[string]any) { delete(b, "scheduledAt") }},
		{"null start", func(b map[string]any) { b["scheduledAt"] = nil }},
		{"zero start", func(b map[string]any) { b["scheduledAt"] = "0001-01-01T00:00:00Z" }},
		{"zero user", func(b map[string]any) { b["userId"] = uuid.Nil.String() }},
		{"missing end", func(b map[string]any) { delete(b, "finishedAt") }},
		{"missing user", func(b map[string]any) { delete(b, "userId") }},
		{"invalid user", func(b map[string]any) { b["userId"] = invalidValue }},
		{"invalid date", func(b map[string]any) { b["scheduledAt"] = "2026-02-30" }},
		{"equal dates", func(b map[string]any) { b["finishedAt"] = b["scheduledAt"] }},
		{"reversed dates", func(b map[string]any) { b["finishedAt"] = "2026-10-18T09:00:00Z" }},
		{"negative notification", func(b map[string]any) { b["notifyBeforeSeconds"] = -1 }},
	}
	for _, tc := range cases {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				// A nil calendar ensures validation rejects the request before any service call.
				s := testServer()
				s.c = nil
				b := eventBody()
				b["id"] = uuid.NewString()
				tc.change(b)
				handler := s.createEvent
				if method == "PUT" {
					handler = s.updateEvent
				}
				call(t, handler, method, "/event", b, 400)
			})
		}
	}
	s := testServer()
	call(t, s.updateEvent, "PUT", "/event", eventBody(), 400)
	for _, id := range []string{"", invalidValue, uuid.Nil.String()} {
		call(t, s.deleteEvent, "DELETE", "/event?id="+id, nil, 400)
	}
	for _, handler := range []http.HandlerFunc{s.getByDay, s.getByWeek, s.getByMonth} {
		for _, date := range []string{"", "2026-02-30", invalidValue} {
			call(t, handler, "GET", "/event?date="+date, nil, 400)
		}
	}
	for _, handler := range []http.HandlerFunc{s.createEvent, s.updateEvent} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequestWithContext(t.Context(), "POST", "/event", bytes.NewBufferString("{")))
		if w.Code != 400 {
			t.Fatalf("malformed JSON: status %d", w.Code)
		}
	}
}

func TestHTTPDateConflicts(t *testing.T) {
	s := testServer()
	body := eventBody()
	call(t, s.createEvent, "POST", "/event", body, 200)
	call(t, s.createEvent, "POST", "/event", body, 409)
	body["scheduledAt"] = "2026-10-18T12:00:00Z"
	body["finishedAt"] = "2026-10-18T13:00:00Z"
	data := call(t, s.createEvent, "POST", "/event", body, 200)
	var created Response
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatal(err)
	}
	body = eventBody()
	body["id"] = created.ID.String()
	call(t, s.updateEvent, "PUT", "/event", body, 409)
}

type unavailableCalendar struct {
	Calendar
	err error
}

func (c unavailableCalendar) CreateEvent(context.Context, event.Event) (event.Event, error) {
	return event.Event{}, c.err
}

func (c unavailableCalendar) UpdateEvent(context.Context, event.Event) (event.Event, error) {
	return event.Event{}, c.err
}
func (c unavailableCalendar) DeleteEvent(context.Context, string) error { return c.err }
func (c unavailableCalendar) ListByDay(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func (c unavailableCalendar) ListByWeek(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func (c unavailableCalendar) ListByMonth(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func TestHTTPServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{fmt.Errorf("wrapped: %w", event.ErrDateBusy), 409},
		{fmt.Errorf("wrapped: %w", event.ErrEventNotFound), 404},
		{errors.New("storage unavailable"), 500},
	} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			s := testServer()
			s.c = unavailableCalendar{err: tc.err}
			b := eventBody()
			b["id"] = uuid.NewString()
			call(t, s.createEvent, "POST", "/event", b, tc.status)
			call(t, s.updateEvent, "PUT", "/event", b, tc.status)
			call(t, s.deleteEvent, "DELETE", "/event?id="+uuid.NewString(), nil, tc.status)
		})
	}
	s := testServer()
	s.c = unavailableCalendar{err: errors.New("storage unavailable")}
	for _, handler := range []http.HandlerFunc{s.getByDay, s.getByWeek, s.getByMonth} {
		call(t, handler, "GET", "/event?date=2026-10-18", nil, 500)
	}
}

func TestHTTPPeriodBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, date, start, end string
		handler                func(*Server) http.HandlerFunc
	}{
		{
			"day", "2026-10-18", "2026-10-18T00:00:00Z", "2026-10-19T00:00:00Z",
			func(s *Server) http.HandlerFunc { return s.getByDay },
		},
		{
			"week", "2026-10-18", "2026-10-12T00:00:00Z", "2026-10-19T00:00:00Z",
			func(s *Server) http.HandlerFunc { return s.getByWeek },
		},
		{
			"month", "2028-02-15", "2028-02-01T00:00:00Z", "2028-03-01T00:00:00Z",
			func(s *Server) http.HandlerFunc { return s.getByMonth },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := testServer()
			start, _ := time.Parse(time.RFC3339, tc.start)
			end, _ := time.Parse(time.RFC3339, tc.end)
			for _, when := range []time.Time{start.Add(-time.Second), start, end.Add(-time.Second), end} {
				b := eventBody()
				b["userId"] = uuid.NewString()
				b["scheduledAt"] = when.Format(time.RFC3339)
				b["finishedAt"] = when.Add(time.Second).Format(time.RFC3339)
				call(t, s.createEvent, "POST", "/event", b, 200)
			}
			data := call(t, tc.handler(s), "GET", "/event?date="+tc.date, nil, 200)
			var events []Response
			if err := json.Unmarshal(data, &events); err != nil {
				t.Fatal(err)
			}
			if len(events) != 2 {
				t.Fatalf("expected 2 events, got %d", len(events))
			}
			for _, e := range events {
				if e.ScheduledAt.Before(start) || !e.ScheduledAt.Before(end) {
					t.Fatalf("event outside period: %+v", e)
				}
			}
		})
	}
}
