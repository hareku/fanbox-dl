package fanbox

import (
	"io"
	"log"
)

type Logger struct {
	L       *log.Logger
	Verbose bool
}

type NewLoggerInput struct {
	Out     io.Writer
	Verbose bool
}

func NewLogger(in *NewLoggerInput) *Logger {
	return &Logger{
		L:       log.New(in.Out, "[fanbox-dl] ", log.LstdFlags),
		Verbose: in.Verbose,
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.L.Printf("[INFO] "+format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.L.Printf("[ERR] "+format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Verbose {
		l.L.Printf("[DEBUG] "+format, v...)
	}
}

// for retryablehttp.LeveledLogger
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Errorf(msg, keysAndValues...)
}
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Infof(msg, keysAndValues...)
}
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Debugf(msg, keysAndValues...)
}
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Debugf(msg, keysAndValues...)
}
