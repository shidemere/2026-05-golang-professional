// Package main is calendar service entrypoint.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/app"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/config"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/logger"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	memorystorage "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/storage/memory"
	internalgrpc "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/grpc/event/generated"
	internalhttp "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/http"
	"google.golang.org/grpc"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/configs/config.yaml", "Path to configuration file")
}

func main() {
	if err := run(); err != nil {
		slog.Error("calendar failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return nil
	}

	config := config.NewConfig(configFile)
	logg := logger.New(config.LoggerConf)

	calendar, closeStorage, err := newCalendar(config, logg)
	if err != nil {
		return err
	}
	defer closeStorage()

	httpServer := internalhttp.NewServer(logg, calendar, config.HTTPServerConf.Host, config.HTTPServerConf.Port)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(internalgrpc.LoggingMiddleware(logg)))
	grpcCalendarServer := internalgrpc.NewServer(logg, calendar)
	generated.RegisterEventServiceServer(grpcServer, grpcCalendarServer)
	grpsPort := fmt.Sprintf(":%v", config.GRPCServerConf.Port)
	var listenConfig net.ListenConfig
	grpclis, err := listenConfig.Listen(context.Background(), "tcp", grpsPort)
	if err != nil {
		return fmt.Errorf("listen for grpc: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("calendar is running...")

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := httpServer.Start(ctx); err != nil {
			logg.Error("failed to start http server: " + err.Error())
			cancel()
		}
	})

	wg.Go(func() {
		if err := grpcServer.Serve(grpclis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logg.Error("faied to start grpc server: " + err.Error())
			cancel()
		}
	})
	<-ctx.Done()

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelTimeout()

	if err := shutdownServers(ctxTimeout, httpServer, grpcServer); err != nil {
		logg.Error("failed to stop servers", "err", err)
	}
	wg.Wait()
	return nil
}

func newCalendar(config config.Config, logg *slog.Logger) (*app.App, func(), error) {
	if config.UseInMemory {
		return app.New(logg, memorystorage.New()), func() {}, nil
	}
	pool, err := CreatePostgresPool(config, logg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}
	return app.New(logg, event.NewRepository(pool)), pool.Close, nil
}

func CreatePostgresPool(config config.Config, logg *slog.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s pool_max_conns=%d pool_min_conns=%d ",
		config.PostgresConf.User,
		config.PostgresConf.Password,
		config.PostgresConf.Host,
		config.PostgresConf.Port,
		config.PostgresConf.DBName,
		config.PostgresConf.PoolMaxConns,
		config.PostgresConf.PoolMinConns,
	)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logg.Error("can't create postgres pool", "err", err)
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		logg.Error("can't connect to postgres", "err", err)
		pool.Close()
		return nil, err
	}

	return pool, nil
}
