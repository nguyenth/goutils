package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"goutils/log"
	"goutils/log/loggers"
)

func main() {
	// Initialize with multiple loggers
	zeroConfig := loggers.DefaultZeroLoggerConfig()
	zeroConfig.Level = log.DEBUG
	zeroLogger := loggers.NewZeroLogger(zeroConfig)

	log.Init(zeroLogger)

	fmt.Println("=== Basic Hook Usage ===")
	basicHookExample()

	fmt.Println("\n=== Production Hook Setup ===")
	productionHookExample()

	fmt.Println("\n=== Development Hook Setup ===")
	developmentHookExample()

	fmt.Println("\n=== Advanced Hook Combinations ===")
	advancedHookExample()
}

// Example 1: Basic hook usage
func basicHookExample() {
	// Clear any existing hooks
	log.GetHookManager().Clear()

	// Add service metadata to all logs
	serviceHook := log.NewFieldHook(map[string]interface{}{
		"service": "my-service",
		"version": "1.0.0",
		"env":     "development",
	})
	log.AddHook(serviceHook)

	// Add metrics tracking
	metricsHook := log.NewMetricsHook()
	log.AddHook(metricsHook)

	ctx := context.Background()
	log.Info(ctx, "Service started successfully")
	log.Warn(ctx, "Configuration missing, using defaults")
	log.Error(ctx, "Failed to connect to database")

	// Show metrics
	counts := metricsHook.GetCounts()
	fmt.Printf("Log counts: INFO=%d, WARN=%d, ERROR=%d\n",
		counts[log.INFO], counts[log.WARN], counts[log.ERROR])
}

// Example 2: Production-ready setup
func productionHookExample() {
	log.GetHookManager().Clear()

	// 1. Environment metadata
	envHook := log.NewFieldHook(map[string]interface{}{
		"environment": "production",
		"deployment":  "v2.1.0",
		"datacenter":  "us-east-1",
		"instance_id": "i-1234567890abcdef0",
	})
	log.AddHook(envHook)

	// 2. Filter out debug logs in production
	productionFilter := log.NewFilterHook(func(entry *log.LogEntry) bool {
		return entry.Level >= log.INFO
	})
	log.AddHook(productionFilter)

	// 3. Add request context data
	requestHook := log.HookFunc(func(entry *log.LogEntry) error {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}

		// Extract data from context
		if traceID := entry.Context.Value("trace_id"); traceID != nil {
			entry.Fields["trace_id"] = traceID
		}
		if userID := entry.Context.Value("user_id"); userID != nil {
			entry.Fields["user_id"] = userID
		}

		return nil
	})
	log.AddHook(requestHook)

	// 4. Error alerting hook
	alertHook := log.HookFunc(func(entry *log.LogEntry) error {
		if entry.Level >= log.ERROR {
			fmt.Printf("🚨 PRODUCTION ALERT: [%s] %s\n", entry.Level, entry.Message)
		}
		return nil
	})
	log.AddHook(alertHook)

	// Test with context
	ctx := context.WithValue(context.Background(), "trace_id", "trace-abc123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	log.Debug(ctx, "Debug message (filtered in production)")
	log.Info(ctx, "User authentication successful")
	log.Error(ctx, "Payment processing failed")
}

// Example 3: Development setup
func developmentHookExample() {
	log.GetHookManager().Clear()

	// 1. Enhanced debugging information
	debugHook := log.HookFunc(func(entry *log.LogEntry) error {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}

		// Add caller information
		if entry.Caller != nil {
			entry.Fields["caller"] = fmt.Sprintf("%s:%d",
				entry.Caller.File, entry.Caller.Line)
		}

		// Add timestamp
		entry.Fields["timestamp"] = entry.Timestamp.Format("15:04:05.000")

		return nil
	})
	log.AddHook(debugHook)

	// 2. Color coding for different levels
	colorHook := log.HookFunc(func(entry *log.LogEntry) error {
		switch entry.Level {
		case log.DEBUG:
			entry.Message = "🔵 DEBUG: " + entry.Message
		case log.INFO:
			entry.Message = "🟢 INFO: " + entry.Message
		case log.WARN:
			entry.Message = "🟡 WARN: " + entry.Message
		case log.ERROR:
			entry.Message = "🔴 ERROR: " + entry.Message
		case log.PANIC:
			entry.Message = "💥 PANIC: " + entry.Message
		}
		return nil
	})
	log.AddHook(colorHook)

	// 3. Performance tracking
	perfHook := &PerformanceTracker{}
	log.AddHook(perfHook)

	ctx := context.Background()
	log.Debug(ctx, "Starting development session")
	log.Info(ctx, "Loading configuration")
	log.Warn(ctx, "Using fallback configuration")
	log.Error(ctx, "Failed to load optional plugin")
}

// Example 4: Advanced combinations
func advancedHookExample() {
	log.GetHookManager().Clear()

	// 1. Sensitive data redaction
	redactionHook := log.HookFunc(func(entry *log.LogEntry) error {
		// Redact sensitive data from message
		sensitivePatterns := []string{"password", "token", "secret", "key"}

		for _, pattern := range sensitivePatterns {
			if strings.Contains(strings.ToLower(entry.Message), pattern) {
				entry.Message = strings.ReplaceAll(entry.Message, pattern, "[REDACTED]")
			}
		}

		// Redact from args
		for i, arg := range entry.Args {
			if str, ok := arg.(string); ok {
				for _, pattern := range sensitivePatterns {
					if strings.Contains(strings.ToLower(str), pattern) {
						entry.Args[i] = "[REDACTED]"
					}
				}
			}
		}

		return nil
	})
	log.AddHook(redactionHook)

	// 2. Rate limiting to prevent log spam
	rateLimitHook := &RateLimitHook{
		maxLogs:    5,
		timeWindow: time.Second,
		logs:       make([]time.Time, 0),
	}
	log.AddHook(rateLimitHook)

	// 3. Structured logging enhancement
	structuredHook := log.HookFunc(func(entry *log.LogEntry) error {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}

		entry.Fields["log_id"] = fmt.Sprintf("log_%d", time.Now().UnixNano())
		entry.Fields["process_id"] = os.Getpid()
		entry.Fields["level_numeric"] = int(entry.Level)

		return nil
	})
	log.AddHook(structuredHook)

	ctx := context.Background()

	// Test scenarios
	log.Info(ctx, "Processing user data")
	log.Warn(ctx, "API token is about to expire: abc123")
	log.Error(ctx, "Database password authentication failed")

	// Test rate limiting
	fmt.Println("Testing rate limiting (should see max 5 logs):")
	for i := 0; i < 10; i++ {
		log.Debug(ctx, "Rapid log entry", "iteration", i)
	}
}

// Custom hook implementations

// PerformanceTracker tracks operation performance
type PerformanceTracker struct {
	operations map[string]time.Time
}

func (pt *PerformanceTracker) BeforeLog(entry *log.LogEntry) error {
	if pt.operations == nil {
		pt.operations = make(map[string]time.Time)
	}

	if strings.Contains(entry.Message, "Starting") {
		pt.operations["current"] = entry.Timestamp
	}

	return nil
}

func (pt *PerformanceTracker) AfterLog(entry *log.LogEntry, err error) {
	if strings.Contains(entry.Message, "Loading") || strings.Contains(entry.Message, "Failed") {
		if startTime, exists := pt.operations["current"]; exists {
			duration := entry.Timestamp.Sub(startTime)
			fmt.Printf("⏱️ Operation took: %v\n", duration)
		}
	}
}

func (pt *PerformanceTracker) GetLevels() []log.Level {
	return nil // All levels
}

// RateLimitHook prevents log spam
type RateLimitHook struct {
	maxLogs    int
	timeWindow time.Duration
	logs       []time.Time
}

func (rlh *RateLimitHook) BeforeLog(entry *log.LogEntry) error {
	now := entry.Timestamp

	// Clean old entries
	cutoff := now.Add(-rlh.timeWindow)
	var newLogs []time.Time
	for _, logTime := range rlh.logs {
		if logTime.After(cutoff) {
			newLogs = append(newLogs, logTime)
		}
	}
	rlh.logs = newLogs

	// Check rate limit
	if len(rlh.logs) >= rlh.maxLogs {
		return log.ErrLogFiltered
	}

	// Add this log
	rlh.logs = append(rlh.logs, now)
	return nil
}

func (rlh *RateLimitHook) AfterLog(entry *log.LogEntry, err error) {}

func (rlh *RateLimitHook) GetLevels() []log.Level {
	return nil
}
