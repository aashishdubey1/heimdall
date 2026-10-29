package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func New(level string) (*slog.Logger, error) {
	var logLevel slog.Level

	switch level {
	case strings.ToLower("info"):
		logLevel = slog.LevelInfo
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("unsupported log level %q", level)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler), nil

}
