package internalgrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingMiddleware logs the method, peer, duration and status of each RPC.
func LoggingMiddleware(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		ip := "unknown"
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			ip = p.Addr.String()
		}
		logger.Info(
			"grpc request handled",
			"ip", ip,
			"time", start,
			"method", info.FullMethod,
			"latency", time.Since(start),
			"status", status.Code(err).String(),
		)
		return resp, err
	}
}
