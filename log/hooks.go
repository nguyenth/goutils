package log

import (
	"context"
	"fmt"
	"time"
)

// LogEntry represents a log entry with all its metadata
type LogEntry struct {
	// Level is the log level
	Level Level
	// Message is the log message
	Message string
	// Args are the message arguments
	Args []interface{}
	// Timestamp is when the log entry was created
	Timestamp time.Time
	// Context is the context passed to the log call
	Context context.Context
	// Fields contains additional structured data
	Fields map[string]interface{}
	// Logger name or source
	Logger string
	// Caller information (file, line, function)
	Caller *CallerInfo
}

// CallerInfo contains information about the caller
type CallerInfo struct {
	File     string
	Line     int
	Function string
}

// Hook interface defines methods that can be called during logging
type Hook interface {
	// BeforeLog is called before the log entry is processed
	// It can modify the log entry or return an error to prevent logging
	BeforeLog(entry *LogEntry) error

	// AfterLog is called after the log entry has been processed
	// It receives the original entry and any error that occurred during logging
	AfterLog(entry *LogEntry, err error)

	// GetLevels returns the log levels this hook should be triggered for
	// Return nil or empty slice to trigger for all levels
	GetLevels() []Level
}

// HookFunc is a function type that implements the Hook interface
type HookFunc func(entry *LogEntry) error

// BeforeLog implements the Hook interface
func (f HookFunc) BeforeLog(entry *LogEntry) error {
	return f(entry)
}

// AfterLog implements the Hook interface (no-op for function hooks)
func (f HookFunc) AfterLog(entry *LogEntry, err error) {
	// No-op for simple function hooks
}

// GetLevels implements the Hook interface (all levels for function hooks)
func (f HookFunc) GetLevels() []Level {
	return nil // nil means all levels
}

// HookManager manages multiple hooks
type HookManager struct {
	hooks []Hook
}

// NewHookManager creates a new hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]Hook, 0),
	}
}

// AddHook adds a hook to the manager
func (hm *HookManager) AddHook(hook Hook) {
	hm.hooks = append(hm.hooks, hook)
}

// RemoveHook removes a hook from the manager
func (hm *HookManager) RemoveHook(hook Hook) {
	for i, h := range hm.hooks {
		if h == hook {
			hm.hooks = append(hm.hooks[:i], hm.hooks[i+1:]...)
			break
		}
	}
}

// ExecuteBeforeHooks executes all before hooks for the given log entry
func (hm *HookManager) ExecuteBeforeHooks(entry *LogEntry) error {
	for _, hook := range hm.hooks {
		// Check if this hook should be triggered for this level
		levels := hook.GetLevels()
		if len(levels) > 0 {
			shouldTrigger := false
			for _, level := range levels {
				if level == entry.Level {
					shouldTrigger = true
					break
				}
			}
			if !shouldTrigger {
				continue
			}
		}

		if err := hook.BeforeLog(entry); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteAfterHooks executes all after hooks for the given log entry
func (hm *HookManager) ExecuteAfterHooks(entry *LogEntry, logErr error) {
	for _, hook := range hm.hooks {
		// Check if this hook should be triggered for this level
		levels := hook.GetLevels()
		if len(levels) > 0 {
			shouldTrigger := false
			for _, level := range levels {
				if level == entry.Level {
					shouldTrigger = true
					break
				}
			}
			if !shouldTrigger {
				continue
			}
		}

		hook.AfterLog(entry, logErr)
	}
}

// GetHooks returns all registered hooks
func (hm *HookManager) GetHooks() []Hook {
	return hm.hooks
}

// Clear removes all hooks
func (hm *HookManager) Clear() {
	hm.hooks = hm.hooks[:0]
}

// Common hook implementations

// FilterHook filters log entries based on a predicate function
type FilterHook struct {
	predicate func(*LogEntry) bool
	levels    []Level
}

// NewFilterHook creates a new filter hook
func NewFilterHook(predicate func(*LogEntry) bool, levels ...Level) *FilterHook {
	return &FilterHook{
		predicate: predicate,
		levels:    levels,
	}
}

// BeforeLog implements the Hook interface
func (fh *FilterHook) BeforeLog(entry *LogEntry) error {
	if !fh.predicate(entry) {
		return ErrLogFiltered
	}
	return nil
}

// AfterLog implements the Hook interface
func (fh *FilterHook) AfterLog(entry *LogEntry, err error) {
	// No-op for filter hooks
}

// GetLevels implements the Hook interface
func (fh *FilterHook) GetLevels() []Level {
	return fh.levels
}

// FieldHook adds additional fields to log entries
type FieldHook struct {
	fields map[string]interface{}
	levels []Level
}

// NewFieldHook creates a new field hook
func NewFieldHook(fields map[string]interface{}, levels ...Level) *FieldHook {
	return &FieldHook{
		fields: fields,
		levels: levels,
	}
}

// BeforeLog implements the Hook interface
func (fh *FieldHook) BeforeLog(entry *LogEntry) error {
	if entry.Fields == nil {
		entry.Fields = make(map[string]interface{})
	}
	for key, value := range fh.fields {
		entry.Fields[key] = value
	}
	return nil
}

// AfterLog implements the Hook interface
func (fh *FieldHook) AfterLog(entry *LogEntry, err error) {
	// No-op for field hooks
}

// GetLevels implements the Hook interface
func (fh *FieldHook) GetLevels() []Level {
	return fh.levels
}

// MetricsHook tracks logging metrics
type MetricsHook struct {
	counters map[Level]int64
	levels   []Level
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook(levels ...Level) *MetricsHook {
	return &MetricsHook{
		counters: make(map[Level]int64),
		levels:   levels,
	}
}

// BeforeLog implements the Hook interface
func (mh *MetricsHook) BeforeLog(entry *LogEntry) error {
	return nil
}

// AfterLog implements the Hook interface
func (mh *MetricsHook) AfterLog(entry *LogEntry, err error) {
	if err == nil {
		mh.counters[entry.Level]++
	}
}

// GetLevels implements the Hook interface
func (mh *MetricsHook) GetLevels() []Level {
	return mh.levels
}

// GetCounts returns the current counts for each level
func (mh *MetricsHook) GetCounts() map[Level]int64 {
	counts := make(map[Level]int64)
	for level, count := range mh.counters {
		counts[level] = count
	}
	return counts
}

// Reset resets all counters
func (mh *MetricsHook) Reset() {
	mh.counters = make(map[Level]int64)
}

// Errors
var (
	// ErrLogFiltered is returned when a log entry is filtered out
	ErrLogFiltered = fmt.Errorf("log entry filtered")
)
