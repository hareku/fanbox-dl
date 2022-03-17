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

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Verbose {
		l.L.Printf("[DEBUG] "+format, v...)
	}
}
