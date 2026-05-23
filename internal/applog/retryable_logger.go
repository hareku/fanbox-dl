package applog

import (
	"log/slog"
	"strings"
)

// RetryableLeveledLogger wraps slog.Logger to implement retryablehttp.LeveledLogger.
// It upgrades "retrying request" debug logs for HTTP 429 responses to Info level.
type RetryableLeveledLogger struct {
	logger *slog.Logger
}

func NewRetryableLeveledLogger(logger *slog.Logger) *RetryableLeveledLogger {
	return &RetryableLeveledLogger{logger: logger}
}

func (l *RetryableLeveledLogger) Error(msg string, keysAndValues ...any) {
	l.logger.Error(msg, keysAndValues...)
}

func (l *RetryableLeveledLogger) Warn(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, keysAndValues...)
}

func (l *RetryableLeveledLogger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *RetryableLeveledLogger) Debug(msg string, keysAndValues ...any) {
	if msg == "retrying request" {
		for i := 1; i < len(keysAndValues); i += 2 {
			if v, ok := keysAndValues[i].(string); ok && strings.Contains(v, "(status: 429)") {
				l.logger.Info(msg, keysAndValues...)
				return
			}
		}
	}
	l.logger.Debug(msg, keysAndValues...)
}
