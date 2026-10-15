// Package main is calendar service entrypoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/app"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/config"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/logger"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
	memorystorage "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/storage/memory"
	internalhttp "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/transport/http"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := config.NewConfig(configFile)
	logg := logger.New(config.LoggerConf)

	var eventRepository *event.Repository
	if !config.UseInMemory {
		pool, err := CreatePostgresPool(config, logg)
		if err != nil {
			logg.Error("got error while trying connect postgres", "err", err)
			os.Exit(1)
		}

		eventRepository = event.NewRepository(pool)
	}

	var calendar *app.App
	if eventRepository == nil {
		calendar = app.New(logg, memorystorage.New())
	} else {
		calendar = app.New(logg, eventRepository)
	}

	server := internalhttp.NewServer(logg, calendar, config.ServerConf.Host, config.ServerConf.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
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
	}

	return pool, nil
}
