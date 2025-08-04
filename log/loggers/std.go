package loggers

import (
	"context"
	stdlog "log"

	"goutils/log"
)

var _ log.Logger = (*StdLogger)(nil)

type StdLogger struct {
	level log.Level
}

func NewStdLogger(level log.Level) *StdLogger {
	return &StdLogger{
		level: level,
	}
}

func (l *StdLogger) Log(ctx context.Context, level log.Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	if len(args) > 0 {
		stdlog.Printf("[%s] %s: %v\n", level, msg, args)
		return
	}

	stdlog.Printf("[%s] %s\n", level, msg)
}

func (l *StdLogger) GetLevel() log.Level {
	return l.level
}
