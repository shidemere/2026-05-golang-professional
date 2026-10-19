package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/app"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	memory "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dayPeriod   = "day"
	weekPeriod  = "week"
	monthPeriod = "month"
	testDate    = "2026-10-18"
)

func testClient(t *testing.T, calendars ...Calendar) (generated.EventServiceClient, context.Context) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	var calendar Calendar = app.New(logger, memory.New())
	if len(calendars) > 0 {
		calendar = calendars[0]
	}
	generated.RegisterEventServiceServer(server, NewServer(logger, calendar))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn, err := grpc.NewClient("passthrough:///calendar",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return generated.NewEventServiceClient(conn), ctx
}

func createRequest() *generated.CreateEventRequst {
	start := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	return &generated.CreateEventRequst{
		Title: "Meeting", ScheduledAt: timestamppb.New(start), FinishedAt: timestamppb.New(start.Add(time.Hour)),
		UserId: "22222222-2222-4222-8222-222222222222", Description: "Agenda", NotifyBeforeSeconds: 60,
	}
}

func updateRequest(id string, in *generated.CreateEventRequst) *generated.UpdateEventRequest {
	return &generated.UpdateEventRequest{
		Id: id, Title: in.Title, ScheduledAt: in.ScheduledAt, FinishedAt: in.FinishedAt,
		UserId: in.UserId, Description: in.Description, NotifyBeforeSeconds: in.NotifyBeforeSeconds,
	}
}

func TestGRPCLifecycle(t *testing.T) {
	client, ctx := testClient(t)
	created, err := client.CreateEvent(ctx, createRequest())
	if err != nil {
		t.Fatal(err)
	}
	if id, err := uuid.Parse(created.Id); err != nil || id == uuid.Nil {
		t.Fatalf("invalid created ID: %s", created.Id)
	}
	req := createRequest()
	req.Title = "Updated"
	req.Description = "Changed"
	req.NotifyBeforeSeconds = 120
	req.UserId = uuid.NewString()
	req.ScheduledAt = timestamppb.New(req.ScheduledAt.AsTime().Add(2 * time.Hour))
	req.FinishedAt = timestamppb.New(req.FinishedAt.AsTime().Add(2 * time.Hour))
	updated, err := client.UpdateEvent(ctx, updateRequest(created.Id, req))
	if err != nil {
		t.Fatal(err)
	}
	assertUpdatedEvent(t, created.Id, updated, req)
	lists := []struct {
		name string
		list func() (*generated.Events, error)
	}{
		{dayPeriod, func() (*generated.Events, error) { return client.ListByDay(ctx, &generated.Day{Day: req.ScheduledAt}) }},
		{weekPeriod, func() (*generated.Events, error) {
			return client.ListByWeek(ctx, &generated.Week{Week: req.ScheduledAt})
		}},
		{monthPeriod, func() (*generated.Events, error) {
			return client.ListByMonth(ctx, &generated.Month{Month: req.ScheduledAt})
		}},
	}

	for _, tc := range lists {
		t.Run(tc.name, func(t *testing.T) {
			events, err := tc.list()
			if err != nil {
				t.Fatal(err)
			}
			assertListedEvent(t, events, created.Id, req)
		})
	}
	if _, err := client.DeleteEvent(ctx, &generated.DeleteEventRequest{Id: created.Id}); err != nil {
		t.Fatal(err)
	}
	events, err := client.ListByDay(ctx, &generated.Day{Day: req.ScheduledAt})
	if err != nil {
		t.Fatal(err)
	}
	if len(events.Event) != 0 {
		t.Fatalf("deleted event still listed: %v", events)
	}
	_, err = client.DeleteEvent(ctx, &generated.DeleteEventRequest{Id: created.Id})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
	if _, err := client.UpdateEvent(ctx, updateRequest(created.Id, req)); status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestGRPCValidation(t *testing.T) {
	client, ctx := testClient(t)
	cases := []struct {
		name   string
		change func(*generated.CreateEventRequst)
	}{
		{"empty title", func(r *generated.CreateEventRequst) { r.Title = "" }},
		{"missing start", func(r *generated.CreateEventRequst) { r.ScheduledAt = nil }},
		{"missing end", func(r *generated.CreateEventRequst) { r.FinishedAt = nil }},
		{"invalid start seconds", func(r *generated.CreateEventRequst) {
			r.ScheduledAt = &timestamppb.Timestamp{Seconds: 253402300800}
		}},
		{"invalid end nanos", func(r *generated.CreateEventRequst) {
			r.FinishedAt = &timestamppb.Timestamp{Seconds: r.FinishedAt.Seconds, Nanos: 1000000000}
		}},
		{"equal dates", func(r *generated.CreateEventRequst) { r.FinishedAt = r.ScheduledAt }},
		{"reversed dates", func(r *generated.CreateEventRequst) {
			r.FinishedAt = timestamppb.New(r.ScheduledAt.AsTime().Add(-time.Hour))
		}},
		{"negative notification", func(r *generated.CreateEventRequst) { r.NotifyBeforeSeconds = -1 }},
		{"missing user", func(r *generated.CreateEventRequst) { r.UserId = "" }},
		{"invalid user", func(r *generated.CreateEventRequst) { r.UserId = "invalid" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			existing, err := client.CreateEvent(ctx, createRequest())
			if err != nil {
				t.Fatal(err)
			}
			req := createRequest()
			tc.change(req)
			if _, err := client.CreateEvent(ctx, req); status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
			if _, err := client.UpdateEvent(ctx, updateRequest(existing.Id, req)); status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
			saved, err := client.ListByDay(ctx, &generated.Day{Day: createRequest().ScheduledAt})
			if err != nil {
				t.Fatal(err)
			}
			if len(saved.Event) != 1 ||
				saved.Event[0].Title != existing.Title ||
				!saved.Event[0].ScheduledAt.AsTime().Equal(existing.ScheduledAt.AsTime()) ||
				!saved.Event[0].FinishedAt.AsTime().Equal(existing.FinishedAt.AsTime()) ||
				saved.Event[0].NotifyBeforeSeconds != existing.NotifyBeforeSeconds ||
				saved.Event[0].UserId != existing.UserId {
				t.Fatalf("invalid update changed storage: %v", saved)
			}
			if _, err := client.DeleteEvent(ctx, &generated.DeleteEventRequest{Id: existing.Id}); err != nil {
				t.Fatal(err)
			}
		})
	}
	// Invalid requests must leave storage untouched.
	events, err := client.ListByDay(ctx, &generated.Day{Day: createRequest().ScheduledAt})
	if err != nil {
		t.Fatal(err)
	}
	if len(events.Event) != 0 {
		t.Fatalf("invalid requests saved events: %v", events)
	}
}

func TestGRPCDateConflicts(t *testing.T) {
	client, ctx := testClient(t)
	req := createRequest()
	if _, err := client.CreateEvent(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateEvent(ctx, req); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}
	next := createRequest()
	next.ScheduledAt = timestamppb.New(next.ScheduledAt.AsTime().Add(2 * time.Hour))
	next.FinishedAt = timestamppb.New(next.FinishedAt.AsTime().Add(2 * time.Hour))
	created, err := client.CreateEvent(ctx, next)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateEvent(ctx, updateRequest(created.Id, req)); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}
}

func TestGRPCPeriodBoundaries(t *testing.T) {
	for _, tc := range []struct{ name, date, start, end string }{
		{dayPeriod, testDate, testDate, "2026-10-19"},
		{weekPeriod, testDate, "2026-10-12", "2026-10-19"},
		{monthPeriod, "2028-02-15", "2028-02-01", "2028-03-01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, ctx := testClient(t)
			start, _ := time.Parse(time.DateOnly, tc.start)
			end, _ := time.Parse(time.DateOnly, tc.end)
			date, _ := time.Parse(time.DateOnly, tc.date)
			for _, when := range []time.Time{start.Add(-time.Second), start, end.Add(-time.Second), end} {
				req := createRequest()
				req.UserId = uuid.NewString()
				req.ScheduledAt = timestamppb.New(when)
				req.FinishedAt = timestamppb.New(when.Add(time.Second))
				if _, err := client.CreateEvent(ctx, req); err != nil {
					t.Fatal(err)
				}
			}
			var events *generated.Events
			var err error
			switch tc.name {
			case dayPeriod:
				events, err = client.ListByDay(ctx, &generated.Day{Day: timestamppb.New(date)})
			case weekPeriod:
				events, err = client.ListByWeek(ctx, &generated.Week{Week: timestamppb.New(date)})
			case monthPeriod:
				events, err = client.ListByMonth(ctx, &generated.Month{Month: timestamppb.New(date)})
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(events.Event) != 2 {
				t.Fatalf("expected 2 events, got %v", events)
			}
			for _, e := range events.Event {
				if e == nil {
					t.Fatal("nil event")
				}
				when := e.ScheduledAt.AsTime()
				if when.Before(start) || !when.Before(end) {
					t.Fatalf("event outside period: %v", e)
				}
			}
		})
	}
}

// errorCalendar returns service failures without involving a real database.
type errorCalendar struct{ err error }

func (c errorCalendar) CreateEvent(context.Context, event.Event) (event.Event, error) {
	return event.Event{}, c.err
}

func (c errorCalendar) UpdateEvent(context.Context, event.Event) (event.Event, error) {
	return event.Event{}, c.err
}
func (c errorCalendar) DeleteEvent(context.Context, string) error { return c.err }
func (c errorCalendar) ListByDay(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func (c errorCalendar) ListByWeek(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func (c errorCalendar) ListByMonth(context.Context, time.Time) ([]event.Event, error) {
	return nil, c.err
}

func TestGRPCServiceErrorCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{"wrapped not found", fmt.Errorf("repository: %w", event.ErrEventNotFound), codes.NotFound},
		{"wrapped conflict", fmt.Errorf("repository: %w", event.ErrDateBusy), codes.AlreadyExists},
		{"storage failure", errors.New("private database connection details"), codes.Internal},
		{"canceled", fmt.Errorf("query: %w", context.Canceled), codes.Canceled},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), codes.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, ctx := testClient(t, errorCalendar{err: tc.err})
			req := createRequest()
			calls := []struct {
				name string
				call func() error
			}{
				{"create", func() error { _, err := client.CreateEvent(ctx, req); return err }},
				{"update", func() error {
					_, err := client.UpdateEvent(ctx, updateRequest(uuid.NewString(), req))
					return err
				}},
				{"delete", func() error {
					_, err := client.DeleteEvent(ctx, &generated.DeleteEventRequest{Id: uuid.NewString()})
					return err
				}},
				{dayPeriod, func() error {
					_, err := client.ListByDay(ctx, &generated.Day{Day: req.ScheduledAt})
					return err
				}},
				{weekPeriod, func() error {
					_, err := client.ListByWeek(ctx, &generated.Week{Week: req.ScheduledAt})
					return err
				}},
				{monthPeriod, func() error {
					_, err := client.ListByMonth(ctx, &generated.Month{Month: req.ScheduledAt})
					return err
				}},
			}
			for _, rpc := range calls {
				t.Run(rpc.name, func(t *testing.T) {
					err := rpc.call()
					if status.Code(err) != tc.code {
						t.Fatalf("expected %v, got %v", tc.code, err)
					}
					if tc.code == codes.Internal && strings.Contains(err.Error(), tc.err.Error()) {
						t.Fatalf("internal details exposed: %v", err)
					}
				})
			}
		})
	}
}

func TestGRPCInvalidTimestamps(t *testing.T) {
	client, ctx := testClient(t)
	for _, stamp := range []*timestamppb.Timestamp{nil, {Seconds: 253402300800}, {Nanos: -1}} {
		if _, err := client.ListByDay(ctx, &generated.Day{Day: stamp}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
		if _, err := client.ListByWeek(ctx, &generated.Week{Week: stamp}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
		if _, err := client.ListByMonth(ctx, &generated.Month{Month: stamp}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	}
}

func assertUpdatedEvent(t *testing.T, id string, updated *generated.Event, req *generated.CreateEventRequst) {
	t.Helper()
	if updated.Id != id ||
		updated.Title != req.Title ||
		updated.UserId != req.UserId ||
		updated.Description != req.Description ||
		updated.NotifyBeforeSeconds != req.NotifyBeforeSeconds ||
		!updated.ScheduledAt.AsTime().Equal(req.ScheduledAt.AsTime()) ||
		!updated.FinishedAt.AsTime().Equal(req.FinishedAt.AsTime()) {
		t.Fatalf("update lost fields: %v", updated)
	}
}

func TestGRPCInvalidIDs(t *testing.T) {
	client, ctx := testClient(t)
	for _, id := range []string{"", "invalid"} {
		t.Run("id/"+id, func(t *testing.T) {
			if _, err := client.UpdateEvent(ctx, updateRequest(id, createRequest())); status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
			_, err := client.DeleteEvent(ctx, &generated.DeleteEventRequest{Id: id})
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
		})
	}
}

func assertListedEvent(t *testing.T, events *generated.Events, id string, req *generated.CreateEventRequst) {
	t.Helper()
	if len(events.Event) != 1 ||
		events.Event[0] == nil ||
		events.Event[0].Id != id ||
		events.Event[0].UserId != req.UserId ||
		events.Event[0].Title != req.Title {
		t.Fatalf("incorrect list: %v", events)
	}
}
