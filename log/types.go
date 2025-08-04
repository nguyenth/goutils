package log

import "context"

type Logger interface {
	Log(ctx context.Context, level Level, msg string, args ...interface{})
	GetLevel() Level
}

// FieldLogger interface for loggers that support structured fields
type FieldLogger interface {
	Logger
	LogWithFields(ctx context.Context, level Level, msg string, fields []Field, args ...interface{})
}
