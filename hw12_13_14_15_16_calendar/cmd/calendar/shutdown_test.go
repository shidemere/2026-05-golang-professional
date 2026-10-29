package main

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type testHTTPShutdown struct{ stop func(context.Context) error }

func (s testHTTPShutdown) Stop(ctx context.Context) error { return s.stop(ctx) }

type testGRPCShutdown struct {
	graceful func()
	forced   func()
}

func (s testGRPCShutdown) GracefulStop() { s.graceful() }
func (s testGRPCShutdown) Stop()         { s.forced() }

func TestShutdownServersStopsBothConcurrently(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	httpStarted := make(chan struct{})
	grpcStarted := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	done := make(chan error, 1)
	go func() {
		done <- shutdownServers(ctx,
			testHTTPShutdown{stop: func(ctx context.Context) error {
				close(httpStarted)
				select {
				case <-grpcStarted:
				case <-ctx.Done():
					return ctx.Err()
				}
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}},
			testGRPCShutdown{
				graceful: func() { close(grpcStarted); <-release },
				forced:   func() { once.Do(func() { close(release) }) },
			},
		)
	}()
	select {
	case <-httpStarted:
	case <-ctx.Done():
		t.Fatal("HTTP shutdown did not start")
	}
	select {
	case <-grpcStarted:
	case <-ctx.Done():
		t.Fatal("gRPC shutdown did not start concurrently")
	}
	select {
	case err := <-done:
		t.Fatalf("shutdown returned before active requests completed: %v", err)
	default:
	}
	once.Do(func() { close(release) })
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("shutdown did not finish")
	}
}

func TestShutdownServersForcesGRPCOnDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	forced := make(chan struct{})
	err := shutdownServers(ctx,
		testHTTPShutdown{stop: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }},
		testGRPCShutdown{graceful: func() { <-forced }, forced: func() { close(forced) }},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	select {
	case <-forced:
	default:
		t.Fatal("gRPC Stop was not called")
	}
}

// pendingRPC simulates a handler that is waiting for external work until canceled.
type pendingRPC struct {
	generated.UnimplementedEventServiceServer
	started chan struct{}
}

func (s *pendingRPC) CreateEvent(ctx context.Context, _ *generated.CreateEventRequest) (*generated.Event, error) {
	close(s.started)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestShutdownServersCancelsActiveGRPC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	handler := &pendingRPC{started: make(chan struct{})}
	generated.RegisterEventServiceServer(server, handler)
	defer server.Stop()
	defer listener.Close()
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	conn, err := grpc.NewClient("passthrough:///calendar",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	response := make(chan error, 1)
	go func() {
		_, err := generated.NewEventServiceClient(conn).CreateEvent(ctx, &generated.CreateEventRequest{})
		response <- err
	}()
	select {
	case <-handler.started:
	case <-ctx.Done():
		t.Fatal("RPC did not start")
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancelShutdown()
	err = shutdownServers(shutdownCtx, testHTTPShutdown{stop: func(context.Context) error { return nil }}, server)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected forced stop, got %v", err)
	}
	select {
	case err := <-response:
		if err == nil {
			t.Fatal("interrupted RPC succeeded")
		}
	case <-ctx.Done():
		t.Fatal("RPC was not canceled")
	}
	select {
	case <-served:
	case <-ctx.Done():
		t.Fatal("Serve did not stop")
	}
}
