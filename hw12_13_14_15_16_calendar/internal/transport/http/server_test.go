package internalhttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/test/bufconn"
)

func TestStopBeforeStart(t *testing.T) {
	s := testServer()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	// A stopped server must not open a listener, even if Start runs later.
	if err := s.Start(ctx); err != nil {
		t.Fatalf("normal shutdown returned an error: %v", err)
	}
}

func TestStartReturnsListenError(t *testing.T) {
	s := testServer()
	s.server.Addr = "127.0.0.1:invalid-port"
	if err := s.Start(context.Background()); err == nil {
		t.Fatal("listen failure was swallowed")
	}
}

type closingListener struct {
	net.Listener
	closed chan struct{}
	once   sync.Once
}

func (l *closingListener) Close() error {
	err := l.Listener.Close()
	l.once.Do(func() { close(l.closed) })
	return err
}

//nolint:funlen,gocognit // Keep shutdown coordination and its channel assertions together in each scenario.
func TestHTTPShutdownWithActiveRequest(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(map[bool]string{false: "graceful", true: "deadline"}[force], func(t *testing.T) {
			s := testServer()
			listener := &closingListener{Listener: bufconn.Listen(1024 * 1024), closed: make(chan struct{})}
			started := make(chan struct{})
			release := make(chan struct{})
			handlerDone := make(chan struct{})
			var once sync.Once
			s.mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
				close(started)
				select {
				case <-release:
				case <-r.Context().Done():
				}
				close(handlerDone)
				_, _ = w.Write([]byte("done"))
			})
			serving := make(chan error, 1)
			go func() { serving <- s.server.Serve(listener) }()
			transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return listener.Listener.(*bufconn.Listener).DialContext(ctx)
			}}
			client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
			defer func() { once.Do(func() { close(release) }); _ = s.server.Close(); transport.CloseIdleConnections() }()
			requested := make(chan error, 1)
			go func() {
				req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://calendar/slow", nil)
				if err != nil {
					requested <- err
					return
				}
				resp, err := client.Do(req)
				if err == nil {
					_, err = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
				requested <- err
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("handler did not start")
			}
			shutdownCtx := ctx
			cancelShutdown := func() {}
			if force {
				shutdownCtx, cancelShutdown = context.WithTimeout(ctx, 30*time.Millisecond)
			}
			defer cancelShutdown()
			stopped := make(chan error, 1)
			go func() { stopped <- s.Stop(shutdownCtx) }()
			select {
			case <-listener.closed:
			case <-ctx.Done():
				t.Fatal("shutdown did not close the listener")
			}
			if !force {
				select {
				case err := <-stopped:
					t.Fatalf("shutdown returned with handler active: %v", err)
				default:
				}
				once.Do(func() { close(release) })
			}
			select {
			case err := <-stopped:
				if force && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected deadline, got %v", err)
				}
				if !force && err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("shutdown hung")
			}
			select {
			case <-handlerDone:
			case <-ctx.Done():
				t.Fatal("active handler did not finish")
			}
			select {
			case err := <-requested:
				if !force && err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("client did not finish")
			}
			select {
			case err := <-serving:
				if !errors.Is(err, http.ErrServerClosed) {
					t.Fatalf("unexpected Serve result: %v", err)
				}
			case <-ctx.Done():
				t.Fatal("Serve did not return")
			}
		})
	}
}
