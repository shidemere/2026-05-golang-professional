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
	internalhttp "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/storage/memory"
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

	// Debug(context.Background(), logg, eventRepository)
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
		// config.PostgresConf.SSLmode,
	)
	// TODO: should I use another context?
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

// func Debug(ctx context.Context, logg *slog.Logger, repository *event.Repository) error {
// 	now := time.Now().Truncate(time.Second)
// 	userID := uuid.New()
//
// 	events := []event.Event{
// 		{
// 			ID:                  uuid.New(),
// 			Title:               "debug: today event",
// 			ScheduledAt:         now.Add(1 * time.Hour),
// 			FinishedAt:          now.Add(2 * time.Hour),
// 			Description:         "created from Debug for GetByDay, GetByWeek and GetByMonth",
// 			UserID:              userID,
// 			NotifyBeforeSeconds: int64(15 * time.Minute / time.Second),
// 		},
// 		{
// 			ID:                  uuid.New(),
// 			Title:               "debug: this week event",
// 			ScheduledAt:         now.AddDate(0, 0, 2),
// 			FinishedAt:          now.AddDate(0, 0, 2).Add(90 * time.Minute),
// 			Description:         "created from Debug for week and month selects",
// 			UserID:              userID,
// 			NotifyBeforeSeconds: int64(30 * time.Minute / time.Second),
// 		},
// 		{
// 			ID:                  uuid.New(),
// 			Title:               "debug: next month event",
// 			ScheduledAt:         now.AddDate(0, 1, 0),
// 			FinishedAt:          now.AddDate(0, 1, 0).Add(1 * time.Hour),
// 			Description:         "created from Debug and then deleted",
// 			UserID:              userID,
// 			NotifyBeforeSeconds: int64(1 * time.Hour / time.Second),
// 		},
// 	}
//
// 	fmt.Println("Debug: Create")
// 	for _, item := range events {
// 		created, err := repository.Create(ctx, logg, item)
// 		if err != nil {
// 			return fmt.Errorf("create event %q: %w", item.Title, err)
// 		}
// 		fmt.Printf("created: id=%s title=%q scheduled_at=%s\n",
// 			created.ID, created.Title, created.ScheduledAt.Format(time.RFC3339))
// 	}
//
// 	fmt.Println("Debug: Update")
// 	events[0].Title = "debug: today event updated"
// 	events[0].Description = "updated from Debug"
// 	updated, err := repository.Update(ctx, *logg, events[0])
// 	if err != nil {
// 		return fmt.Errorf("update event %s: %w", events[0].ID, err)
// 	}
// 	fmt.Printf("updated: id=%s title=%q description=%q\n", updated.ID, updated.Title, updated.Description)
//
// 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	startOfNextDay := startOfDay.AddDate(0, 0, 1)
//
// 	fmt.Println("Debug: GetByDay")
// 	dayEvents, err := repository.GetByDay(ctx, *logg, startOfDay, startOfNextDay)
// 	if err != nil {
// 		return fmt.Errorf("get by day: %w", err)
// 	}
// 	fmt.Printf("events by day: %d\n", len(dayEvents))
//
// 	weekday := int(now.Weekday())
// 	if weekday == 0 {
// 		weekday = 7
// 	}
// 	startOfWeek := startOfDay.AddDate(0, 0, 1-weekday)
// 	endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Nanosecond)
//
// 	fmt.Println("Debug: GetByWeek")
// 	weekEvents, err := repository.GetByWeek(ctx, *logg, startOfWeek, endOfWeek)
// 	if err != nil {
// 		return fmt.Errorf("get by week: %w", err)
// 	}
// 	fmt.Printf("events by week: %d\n", len(weekEvents))
//
// 	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
// 	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
//
// 	fmt.Println("Debug: GetByMonth")
// 	monthEvents, err := repository.GetByMonth(ctx, *logg, startOfMonth, endOfMonth)
// 	if err != nil {
// 		return fmt.Errorf("get by month: %w", err)
// 	}
// 	fmt.Printf("events by month: %d\n", len(monthEvents))
//
// 	fmt.Println("Debug: Delete")
// 	if err := repository.Delete(ctx, logg, events[2].ID); err != nil {
// 		return fmt.Errorf("delete event %s: %w", events[2].ID, err)
// 	}
// 	fmt.Printf("deleted: id=%s\n", events[2].ID)
//
// 	return nil
// }
