// Package logger contain logger construction.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/config"
)

// New creates a logger from config.
func New(config config.LoggerConfig) *slog.Logger {
	opts := slog.HandlerOptions{
		Level:     mapLevel(config.Level),
		AddSource: config.AddSource,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	return logger
}

func mapLevel(level string) slog.Level {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Fprintf(os.Stderr, "can't parse logger level, using INFO by default\n")
		return slog.LevelInfo
	}
}
