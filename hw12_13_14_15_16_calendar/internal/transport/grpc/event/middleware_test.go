package internalgrpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type recordingLogger struct{ fields map[string]any }

func (l *recordingLogger) Info(_ string, args ...any) {
	l.fields = make(map[string]any)
	for i := 0; i+1 < len(args); i += 2 {
		l.fields[args[i].(string)] = args[i+1]
	}
}
func (l *recordingLogger) Warn(string, ...any) {}

func TestLoggingMiddlewareResult(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{"success", nil, codes.OK},
		{"failure", status.Error(codes.NotFound, "missing"), codes.NotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := &recordingLogger{}
			ctx := peer.NewContext(context.Background(), &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234},
			})
			response := new(int)
			got, err := LoggingMiddleware(logger)(ctx, nil,
				&grpc.UnaryServerInfo{FullMethod: "/event.EventService/ListByDay"},
				func(context.Context, any) (any, error) { return response, tc.err },
			)
			if got != response || !errors.Is(err, tc.err) {
				t.Fatal("middleware changed handler result")
			}
			if logger.fields["status"] != tc.code.String() ||
				logger.fields["method"] != "/event.EventService/ListByDay" ||
				logger.fields["ip"] != "127.0.0.1:1234" {
				t.Fatalf("incorrect log: %v", logger.fields)
			}
		})
	}
	logger := &recordingLogger{}
	_, _ = LoggingMiddleware(logger)(context.Background(), nil, &grpc.UnaryServerInfo{},
		func(context.Context, any) (any, error) { return nil, nil },
	)
	if logger.fields["ip"] != "unknown" {
		t.Fatalf("missing peer: %v", logger.fields)
	}
}
