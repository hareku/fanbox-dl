package fanbox

import (
	"io"
	"log"
)

type Logger interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

type logger struct {
	l       *log.Logger
	verbose bool
}

type NewLoggerInput struct {
	Out     io.Writer
	Verbose bool
}

func NewLogger(in *NewLoggerInput) Logger {
	return &logger{
		l:       log.New(in.Out, "[fanbox-dl] ", log.LstdFlags),
		verbose: in.Verbose,
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.l.Printf("[INFO] "+format, v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if l.verbose {
		l.l.Printf("[DEBUG] "+format, v...)
	}
}
