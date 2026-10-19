package main

import (
	"context"
	"errors"
	"fmt"
)

type httpShutdownServer interface {
	Stop(context.Context) error
}

type grpcShutdownServer interface {
	GracefulStop()
	Stop()
}

// shutdownServers stops accepting work on both transports concurrently and
// bounds the graceful phase with a shared deadline.
func shutdownServers(ctx context.Context, httpServer httpShutdownServer, grpcServer grpcShutdownServer) error {
	httpDone := make(chan error, 1)
	go func() { httpDone <- httpServer.Stop(ctx) }()

	grpcDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcDone)
	}()

	var grpcErr error
	select {
	case <-grpcDone:
	case <-ctx.Done():
		grpcServer.Stop()
		grpcErr = fmt.Errorf("gRPC shutdown: %w", ctx.Err())
	}
	// HTTP Stop honors the same deadline and closes active connections on timeout.
	httpErr := <-httpDone
	return errors.Join(httpErr, grpcErr)
}
