package log

import "time"

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

const (
	FieldUserID      = "user_id"
	FieldRequestID   = "request_id"
	FieldSessionID   = "session_id"
	FieldTraceID     = "trace_id"
	FieldSpanID      = "span_id"
	FieldService     = "service"
	FieldComponent   = "component"
	FieldMethod      = "method"
	FieldUrl         = "url"
	FieldStatusCode  = "status_code"
	FieldLatency     = "latency"
	FieldError       = "error"
	FieldStack       = "stack"
	FieldVersion     = "version"
	FieldEnvironment = "environment"
	FieldDeviceId    = "device_id"
)

// TypedField represents a field with a specific type for optimized logging
type TypedField struct {
	Key   string
	Value interface{}
}

// Convenience functions for creating fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func WithUserId(value string) Field {
	return Field{Key: FieldUserID, Value: value}
}

func WithRequestId(value string) Field {
	return Field{Key: FieldRequestID, Value: value}
}

func WithSessionId(value string) Field {
	return Field{Key: FieldSessionID, Value: value}
}

func WithTraceId(value string) Field {
	return Field{Key: FieldTraceID, Value: value}
}

func WithSpanId(value string) Field {
	return Field{Key: FieldSpanID, Value: value}
}

func WithService(value string) Field {
	return Field{Key: FieldService, Value: value}
}

func WithComponent(value string) Field {
	return Field{Key: FieldComponent, Value: value}
}

func WithMethod(value string) Field {
	return Field{Key: FieldMethod, Value: value}
}

func WithUrl(value string) Field {
	return Field{Key: FieldUrl, Value: value}
}

func WithStatusCode(value int) Field {
	return Field{Key: FieldStatusCode, Value: value}
}

func WithLatency(value time.Duration) Field {
	return Field{Key: FieldLatency, Value: value}
}

func WithError(value error) Field {
	return Field{Key: FieldError, Value: value}
}

func WithStack(value string) Field {
	return Field{Key: FieldStack, Value: value}
}

func WithVersion(value string) Field {
	return Field{Key: FieldVersion, Value: value}
}

func WithEnvironment(value string) Field {
	return Field{Key: FieldEnvironment, Value: value}
}

func WithDeviceId(value string) Field {
	return Field{Key: FieldDeviceId, Value: value}
}

// separateFieldsAndArgs separates Field items from regular args
func separateFieldsAndArgs(args []interface{}) ([]Field, []interface{}) {
	var fields []Field
	var regularArgs []interface{}

	for _, arg := range args {
		if field, ok := arg.(Field); ok {
			fields = append(fields, field)
		} else {
			regularArgs = append(regularArgs, arg)
		}
	}

	return fields, regularArgs
}
