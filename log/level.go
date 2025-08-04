package log

// These are the integer logging levels used by the logger
type Level int

const (
	DEBUG Level = iota
	TRACE
	INFO
	WARN
	ERROR
	PANIC
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case PANIC:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}
