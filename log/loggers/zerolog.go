package loggers

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"goutils/log"

	"github.com/rs/zerolog"
)

// Ensure ZeroLogger implements both Logger and FieldLogger interfaces
var _ log.Logger = (*ZeroLogger)(nil)
var _ log.FieldLogger = (*ZeroLogger)(nil)

// ZeroLoggerConfig contains configuration options for ZeroLogger
type ZeroLoggerConfig struct {
	// Level is the minimum logging level
	Level log.Level
	// Output is the writer where logs will be written (default: os.Stdout)
	Output io.Writer
	// UseConsoleWriter enables human-readable console output (default: true)
	UseConsoleWriter bool
	// TimeFormat sets the time format for timestamps (default: time.RFC3339)
	TimeFormat string
	// NoColor disables colored output when using console writer
	NoColor bool
	// CallerEnabled adds caller information to log entries
	CallerEnabled bool
}

// DefaultZeroLoggerConfig returns a default configuration
func DefaultZeroLoggerConfig() *ZeroLoggerConfig {
	return &ZeroLoggerConfig{
		Level:            log.INFO,
		Output:           os.Stdout,
		UseConsoleWriter: true,
		TimeFormat:       time.RFC3339,
		NoColor:          false,
		CallerEnabled:    false,
	}
}

// ZeroLogger is a wrapper around zerolog that implements the Logger interface
type ZeroLogger struct {
	logger zerolog.Logger
	level  log.Level
}

// NewZeroLogger creates a new ZeroLogger instance with the given configuration
func NewZeroLogger(config *ZeroLoggerConfig) *ZeroLogger {
	if config == nil {
		config = DefaultZeroLoggerConfig()
	}

	var logger zerolog.Logger

	// Configure output writer
	var writer io.Writer = config.Output
	if config.UseConsoleWriter {
		consoleWriter := zerolog.ConsoleWriter{
			Out:     config.Output,
			NoColor: config.NoColor,
		}
		if config.TimeFormat != "" {
			consoleWriter.TimeFormat = config.TimeFormat
		}
		writer = consoleWriter
	}

	// Create the zerolog logger
	logger = zerolog.New(writer).With().Timestamp().Logger()

	// Add caller information if enabled
	if config.CallerEnabled {
		logger = logger.With().Caller().Logger()
	}

	// Set the zerolog level based on our custom level
	logger = logger.Level(convertToZerologLevel(config.Level))

	return &ZeroLogger{
		logger: logger,
		level:  config.Level,
	}
}

// Log implements the Logger interface
func (z *ZeroLogger) Log(ctx context.Context, level log.Level, msg string, args ...interface{}) {
	// Convert our level to zerolog level and log accordingly
	switch level {
	case log.DEBUG:
		z.logger.Debug().Msgf(msg, args...)
	case log.TRACE:
		z.logger.Trace().Msgf(msg, args...)
	case log.INFO:
		z.logger.Info().Msgf(msg, args...)
	case log.WARN:
		z.logger.Warn().Msgf(msg, args...)
	case log.ERROR:
		z.logger.Error().Msgf(msg, args...)
	case log.PANIC:
		z.logger.Panic().Msgf(msg, args...)
	default:
		z.logger.Info().Msgf(msg, args...)
	}
}

// LogWithFields implements the FieldLogger interface
func (z *ZeroLogger) LogWithFields(ctx context.Context, level log.Level, msg string, fields []log.Field, args ...interface{}) {
	// Get the appropriate zerolog event based on level
	var event *zerolog.Event
	switch level {
	case log.DEBUG:
		event = z.logger.Debug()
	case log.TRACE:
		event = z.logger.Trace()
	case log.INFO:
		event = z.logger.Info()
	case log.WARN:
		event = z.logger.Warn()
	case log.ERROR:
		event = z.logger.Error()
	case log.PANIC:
		event = z.logger.Panic()
	default:
		event = z.logger.Info()
	}

	// Add fields to the event
	for _, field := range fields {
		z.addFieldToEvent(event, field)
	}

	// Add context information if available
	if ctx != nil {
		// Extract common context values
		if requestID := ctx.Value("request_id"); requestID != nil {
			event = event.Str("request_id", fmt.Sprintf("%v", requestID))
		}
		if traceID := ctx.Value("trace_id"); traceID != nil {
			event = event.Str("trace_id", fmt.Sprintf("%v", traceID))
		}
		if userID := ctx.Value("user_id"); userID != nil {
			event = event.Str("user_id", fmt.Sprintf("%v", userID))
		}
	}

	// Log the message with or without formatting
	if len(args) > 0 {
		event.Msgf(msg, args...)
	} else {
		event.Msg(msg)
	}
}

// addFieldToEvent adds a field to a zerolog event with appropriate type handling
func (z *ZeroLogger) addFieldToEvent(event *zerolog.Event, field log.Field) {
	switch v := field.Value.(type) {
	case string:
		event.Str(field.Key, v)
	case int:
		event.Int(field.Key, v)
	case int8:
		event.Int8(field.Key, v)
	case int16:
		event.Int16(field.Key, v)
	case int32:
		event.Int32(field.Key, v)
	case int64:
		event.Int64(field.Key, v)
	case uint:
		event.Uint(field.Key, v)
	case uint8:
		event.Uint8(field.Key, v)
	case uint16:
		event.Uint16(field.Key, v)
	case uint32:
		event.Uint32(field.Key, v)
	case uint64:
		event.Uint64(field.Key, v)
	case float32:
		event.Float32(field.Key, v)
	case float64:
		event.Float64(field.Key, v)
	case bool:
		event.Bool(field.Key, v)
	case time.Time:
		event.Time(field.Key, v)
	case time.Duration:
		event.Dur(field.Key, v)
	case error:
		if v != nil {
			event.Err(v)
		} else {
			event.Str(field.Key, "")
		}
	case []byte:
		event.Bytes(field.Key, v)
	case nil:
		event.Str(field.Key, "")
	default:
		// For complex types, use interface{} which zerolog will JSON encode
		event.Interface(field.Key, v)
	}
}

// GetLevel implements the Logger interface
func (z *ZeroLogger) GetLevel() log.Level {
	return z.level
}

// convertToZerologLevel converts our custom Level to zerolog.Level
func convertToZerologLevel(level log.Level) zerolog.Level {
	switch level {
	case log.DEBUG:
		return zerolog.DebugLevel
	case log.TRACE:
		return zerolog.TraceLevel
	case log.INFO:
		return zerolog.InfoLevel
	case log.WARN:
		return zerolog.WarnLevel
	case log.ERROR:
		return zerolog.ErrorLevel
	case log.PANIC:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
