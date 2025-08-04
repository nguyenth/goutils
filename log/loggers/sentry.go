package loggers

import (
	"context"
	"fmt"
	"time"

	"goutils/log"

	"github.com/getsentry/sentry-go"
)

// Ensure SentryLogger implements both Logger and FieldLogger interfaces
var _ log.Logger = (*SentryLogger)(nil)
var _ log.FieldLogger = (*SentryLogger)(nil)

// SentryLoggerConfig contains configuration options for SentryLogger
type SentryLoggerConfig struct {
	// Level is the minimum logging level
	Level log.Level
	// DSN is the Sentry Data Source Name
	DSN string
	// Environment sets the environment (e.g., "production", "development")
	Environment string
	// ServerName sets the server name
	ServerName string
	// Release sets the release version
	Release string
	// SampleRate for error sampling (0.0 to 1.0)
	SampleRate float64
	// Debug enables Sentry debug mode
	Debug bool
	// AttachStacktrace enables stack traces for all events
	AttachStacktrace bool
	// SendDefaultPII enables sending personally identifiable information
	SendDefaultPII bool
	// FlushTimeout for flushing events to Sentry
	FlushTimeout time.Duration
}

// DefaultSentryLoggerConfig returns a default configuration
func DefaultSentryLoggerConfig(dsn string) *SentryLoggerConfig {
	return &SentryLoggerConfig{
		Level:            log.ERROR, // Only send ERROR and PANIC to Sentry by default
		DSN:              dsn,
		Environment:      "development",
		ServerName:       "",
		Release:          "",
		SampleRate:       1.0,
		Debug:            false,
		AttachStacktrace: true,
		SendDefaultPII:   false,
		FlushTimeout:     2 * time.Second,
	}
}

// SentryLogger is a wrapper around Sentry that implements the Logger interface
type SentryLogger struct {
	hub          *sentry.Hub
	level        log.Level
	flushTimeout time.Duration
}

// NewSentryLogger creates a new SentryLogger instance with the given configuration
func NewSentryLogger(config *SentryLoggerConfig) (*SentryLogger, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		ServerName:       config.ServerName,
		Release:          config.Release,
		SampleRate:       config.SampleRate,
		Debug:            config.Debug,
		AttachStacktrace: config.AttachStacktrace,
		SendDefaultPII:   config.SendDefaultPII,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	return &SentryLogger{
		hub:          sentry.CurrentHub(),
		level:        config.Level,
		flushTimeout: config.FlushTimeout,
	}, nil
}

// NewSentryLoggerSimple creates a new SentryLogger instance with minimal configuration
func NewSentryLoggerSimple(dsn string, level log.Level) (*SentryLogger, error) {
	config := DefaultSentryLoggerConfig(dsn)
	config.Level = level
	return NewSentryLogger(config)
}

// Log implements the Logger interface
func (s *SentryLogger) Log(ctx context.Context, level log.Level, msg string, args ...interface{}) {
	// Only log if the level is at or above our configured level
	if level < s.level {
		return
	}

	// Format the message
	message := fmt.Sprintf(msg, args...)

	// Create a new scope for this log entry
	s.hub.WithScope(func(scope *sentry.Scope) {
		// Set the level
		scope.SetLevel(convertToSentryLevel(level))

		// Add context if available
		if ctx != nil {
			// Add any context values as extra data
			scope.SetContext("log_context", map[string]interface{}{
				"level":   level.String(),
				"message": message,
			})
		}

		// Send different event types based on log level
		switch level {
		case log.ERROR, log.PANIC:
			// Send as exception for errors and panics
			s.hub.CaptureException(fmt.Errorf("%s", message))
		default:
			// Send as message for other levels
			s.hub.CaptureMessage(message)
		}
	})

	// For PANIC level, also flush immediately
	if level == log.PANIC {
		sentry.Flush(s.flushTimeout)
	}
}

// LogWithFields implements the FieldLogger interface
func (s *SentryLogger) LogWithFields(ctx context.Context, level log.Level, msg string, fields []log.Field, args ...interface{}) {
	// Only log if the level is at or above our configured level
	if level < s.level {
		return
	}

	// Format the message with args
	message := msg
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	}

	// Create a new scope for this log entry
	s.hub.WithScope(func(scope *sentry.Scope) {
		// Set the level
		scope.SetLevel(convertToSentryLevel(level))

		// Add fields as tags and extra data
		for _, field := range fields {
			switch field.Key {
			case log.FieldUserID:
				// Special handling for user ID - set as user context
				if userID, ok := field.Value.(string); ok {
					scope.SetUser(sentry.User{ID: userID})
				} else {
					scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", field.Value)})
				}
			case log.FieldRequestID, log.FieldTraceID, log.FieldSessionID:
				// Set important IDs as tags for better filtering
				scope.SetTag(field.Key, fmt.Sprintf("%v", field.Value))
			case log.FieldError:
				// Special handling for errors
				if err, ok := field.Value.(error); ok {
					scope.SetExtra(field.Key, err.Error())
				} else {
					scope.SetExtra(field.Key, field.Value)
				}
			default:
				// Convert field value to appropriate Sentry data
				switch v := field.Value.(type) {
				case string, int, int64, float64, bool:
					// Simple types can be tags for filtering
					if len(fmt.Sprintf("%v", v)) < 200 { // Sentry tag limit
						scope.SetTag(field.Key, fmt.Sprintf("%v", v))
					} else {
						scope.SetExtra(field.Key, v)
					}
				case time.Time:
					// Format time as string
					scope.SetExtra(field.Key, v.Format(time.RFC3339))
				case time.Duration:
					// Convert duration to milliseconds for easier analysis
					scope.SetExtra(field.Key+"_ms", v.Milliseconds())
					scope.SetExtra(field.Key, v.String())
				default:
					// Complex types go to extra data
					scope.SetExtra(field.Key, v)
				}
			}
		}

		// Add context if available
		if ctx != nil {
			// Add any context values as extra data
			scope.SetContext("log_context", map[string]interface{}{
				"level":       level.String(),
				"message":     message,
				"field_count": len(fields),
			})

			// Extract common context values
			if requestID := ctx.Value("request_id"); requestID != nil {
				scope.SetTag("request_id", fmt.Sprintf("%v", requestID))
			}
			if traceID := ctx.Value("trace_id"); traceID != nil {
				scope.SetTag("trace_id", fmt.Sprintf("%v", traceID))
			}
			if userID := ctx.Value("user_id"); userID != nil {
				scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", userID)})
			}
		}

		// Create fingerprint for grouping based on message and key fields
		fingerprint := []string{message}
		for _, field := range fields {
			if field.Key == "component" || field.Key == "service" || field.Key == "method" {
				fingerprint = append(fingerprint, fmt.Sprintf("%s:%v", field.Key, field.Value))
			}
		}
		scope.SetFingerprint(fingerprint)

		// Send different event types based on log level
		switch level {
		case log.ERROR, log.PANIC:
			// Send as exception for errors and panics
			s.hub.CaptureException(fmt.Errorf("%s", message))
		default:
			// Send as message for other levels
			s.hub.CaptureMessage(message)
		}
	})

	// For PANIC level, also flush immediately
	if level == log.PANIC {
		sentry.Flush(s.flushTimeout)
	}
}

// GetLevel implements the Logger interface
func (s *SentryLogger) GetLevel() log.Level {
	return s.level
}

// SetLevel allows changing the logging level
func (s *SentryLogger) SetLevel(level log.Level) {
	s.level = level
}

// Flush flushes any pending events to Sentry
func (s *SentryLogger) Flush() {
	sentry.Flush(s.flushTimeout)
}

// Close closes the Sentry client
func (s *SentryLogger) Close() {
	sentry.Flush(s.flushTimeout)
}

// AddTag adds a tag to all future events
func (s *SentryLogger) AddTag(key, value string) {
	s.hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// AddUser sets user information for future events
func (s *SentryLogger) AddUser(user sentry.User) {
	s.hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(user)
	})
}

// AddExtra adds extra data to all future events
func (s *SentryLogger) AddExtra(key string, value interface{}) {
	s.hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtra(key, value)
	})
}

// AddBreadcrumb adds a breadcrumb for debugging context
func (s *SentryLogger) AddBreadcrumb(message, category string, level sentry.Level, data map[string]interface{}) {
	s.hub.AddBreadcrumb(&sentry.Breadcrumb{
		Message:  message,
		Category: category,
		Level:    level,
		Data:     data,
	}, nil)
}

// convertToSentryLevel converts our custom Level to sentry.Level
func convertToSentryLevel(level log.Level) sentry.Level {
	switch level {
	case log.DEBUG:
		return sentry.LevelDebug
	case log.TRACE:
		return sentry.LevelDebug // Sentry doesn't have trace, use debug
	case log.INFO:
		return sentry.LevelInfo
	case log.WARN:
		return sentry.LevelWarning
	case log.ERROR:
		return sentry.LevelError
	case log.PANIC:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}
