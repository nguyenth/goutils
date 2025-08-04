package log

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	loggers     []Logger = []Logger{}
	hookManager *HookManager
	once        sync.Once
	mu          sync.RWMutex
)

// Init initializes the logger with the given loggers
func Init(logs ...Logger) {
	if len(logs) == 0 {
		return
	}

	once.Do(func() {
		loggers = logs
		hookManager = NewHookManager()
	})
}

// AddHook adds a hook to the global hook manager
func AddHook(hook Hook) {
	mu.Lock()
	defer mu.Unlock()

	if hookManager == nil {
		hookManager = NewHookManager()
	}
	hookManager.AddHook(hook)
}

// RemoveHook removes a hook from the global hook manager
func RemoveHook(hook Hook) {
	mu.Lock()
	defer mu.Unlock()

	if hookManager != nil {
		hookManager.RemoveHook(hook)
	}
}

// GetHookManager returns the global hook manager
func GetHookManager() *HookManager {
	mu.RLock()
	defer mu.RUnlock()
	return hookManager
}

func Debug(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, DEBUG, msg, args...)
}

func Trace(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, TRACE, msg, args...)
}

func Info(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, INFO, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, WARN, msg, args...)
}

func Error(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, ERROR, msg, args...)
}

func Panic(ctx context.Context, msg string, args ...interface{}) {
	logMe(ctx, PANIC, msg, args...)
}

func logMe(ctx context.Context, level Level, msg string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}

	// separate fields from regular args
	fields, regularArgs := separateFieldsAndArgs(args)

	// Create log entry
	entry := &LogEntry{
		Level:     level,
		Message:   msg,
		Args:      regularArgs,
		Timestamp: time.Now(),
		Context:   ctx,
		Fields:    make(map[string]interface{}),
		Caller:    getCaller(),
	}

	// Add fields to entry
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	// Execute before hooks
	mu.RLock()
	hm := hookManager
	mu.RUnlock()

	if hm != nil {
		if err := hm.ExecuteBeforeHooks(entry); err != nil {
			// Log was filtered or blocked by hook
			hm.ExecuteAfterHooks(entry, err)
			return
		}
	}

	// Log to all configured loggers
	var logErr error
	for _, logger := range loggers {
		if level < logger.GetLevel() {
			continue
		}

		if fieldLogger, ok := logger.(FieldLogger); ok {
			// Use structured logging
			err := safeLogWithFields(fieldLogger, ctx, level, entry.Message, fields, entry.Args...)
			if err != nil && logErr == nil {
				logErr = err
			}
		} else {

			// Use the potentially modified message and args from hooks
			finalMsg := entry.Message
			finalArgs := entry.Args

			// If hooks added fields, try to incorporate them
			if len(entry.Fields) > 0 {
				// For simple loggers, append fields to the message
				var fieldsStr []string
				for k, v := range entry.Fields {
					fieldsStr = append(fieldsStr, fmt.Sprintf("%s=%v", k, v))
				}
				if len(fieldsStr) > 0 {
					finalMsg = fmt.Sprintf("%s [%s]", finalMsg, strings.Join(fieldsStr, " "))
				}
			}

			err := safeLog(logger, ctx, level, finalMsg, finalArgs...)
			if err != nil && logErr == nil {
				logErr = err
			}
		}
	}

	// Execute after hooks
	if hm != nil {
		hm.ExecuteAfterHooks(entry, logErr)
	}
}

// safeLog wraps logger.Log to catch panics
func safeLog(logger Logger, ctx context.Context, level Level, msg string, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("logger panic: %v", r)
		}
	}()

	logger.Log(ctx, level, msg, args...)
	return nil
}

// safeLogWithFields wraps logger.LogWithFields to catch panics
func safeLogWithFields(logger FieldLogger, ctx context.Context, level Level, msg string, fields []Field, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("logger panic: %v", r)
		}
	}()

	logger.LogWithFields(ctx, level, msg, fields, args...)
	return nil
}

// getCaller returns caller information
func getCaller() *CallerInfo {
	// Skip: getCaller, logMe, and the public log function
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return nil
	}

	// Get function name
	pc, _, _, ok := runtime.Caller(3)
	var funcName string
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			funcName = fn.Name()
		}
	}

	return &CallerInfo{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}
