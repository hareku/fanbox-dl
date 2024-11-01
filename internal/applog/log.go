package applog

import (
	"log/slog"
	"os"
)

func InitLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	var h slog.Handler
	h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	h = NewContextValueLogHandler(h)

	logger := slog.New(h)
	slog.SetDefault(logger)
}
