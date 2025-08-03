package main

import (
	"context"
	"fmt"

	"goutils/log"
	"goutils/log/loggers"
)

func main() {
	// Initialize logger with standard logger
	stdLogger := loggers.NewStdLogger(log.DEBUG)
	log.Init(stdLogger)

	fmt.Println("=== Simple Logging Example ===")

	// Create a context
	ctx := context.Background()

	// Basic logging at different levels
	log.Debug(ctx, "This is a debug message")
	log.Info(ctx, "Application started successfully")
	log.Warn(ctx, "This is a warning message")
	log.Error(ctx, "This is an error message")

	// Logging with arguments (formatted messages)
	log.Info(ctx, "User %s logged in with ID %d", "john_doe", 12345)
	log.Warn(ctx, "Database connection pool at %d%% capacity", 85)
	log.Error(ctx, "Failed to process order %s for user %d", "order-abc123", 67890)

	// Logging with key-value pairs as arguments
	log.Info(ctx, "Order processed",
		"order_id", "order-xyz789",
		"user_id", 54321,
		"amount", 99.99,
		"status", "completed")

	log.Error(ctx, "Payment failed",
		"order_id", "order-fail123",
		"error_code", "CARD_DECLINED",
		"retry_count", 3)

	// Different logger types example
	fmt.Println("\n=== Using Multiple Loggers ===")

	// Add zerolog for better structured output
	zeroConfig := loggers.DefaultZeroLoggerConfig()
	zeroConfig.Level = log.DEBUG
	zeroLogger := loggers.NewZeroLogger(zeroConfig)

	// Re-initialize with both loggers
	log.Init(stdLogger, zeroLogger)

	log.Info(ctx, "Using multiple loggers now")
	log.Error(ctx, "This appears in both standard and zerolog format",
		"component", "payment_service",
		"transaction_id", "txn_456789")

	// Example of different log levels
	fmt.Println("\n=== Log Level Examples ===")

	// These will all show because we set DEBUG level
	log.Debug(ctx, "Debugging user authentication flow")
	log.Trace(ctx, "Tracing database query execution")
	log.Info(ctx, "User session created")
	log.Warn(ctx, "Session will expire in 5 minutes")
	log.Error(ctx, "Session cleanup failed")

	fmt.Println("\n=== Real-world Usage Examples ===")

	// Simulate real application scenarios
	simulateUserLogin()
	simulateOrderProcessing()
	simulateErrorHandling()
}

func simulateUserLogin() {
	ctx := context.Background()

	log.Info(ctx, "User login attempt", "username", "alice@example.com")
	log.Debug(ctx, "Validating user credentials")
	log.Info(ctx, "User authentication successful",
		"user_id", 123,
		"username", "alice@example.com",
		"login_time", "2025-08-03T10:30:00Z")
}

func simulateOrderProcessing() {
	ctx := context.Background()

	log.Info(ctx, "Processing new order", "order_id", "ORD-2025-001")
	log.Debug(ctx, "Validating inventory", "product_id", "PROD-456", "quantity", 2)
	log.Info(ctx, "Inventory check passed")
	log.Debug(ctx, "Calculating total amount", "subtotal", 89.98, "tax", 7.20, "total", 97.18)
	log.Info(ctx, "Order completed successfully",
		"order_id", "ORD-2025-001",
		"total_amount", 97.18,
		"payment_method", "credit_card")
}

func simulateErrorHandling() {
	ctx := context.Background()

	log.Error(ctx, "Database connection failed",
		"database", "orders_db",
		"error", "connection timeout",
		"retry_attempt", 1)

	log.Warn(ctx, "Falling back to read replica", "replica_id", "replica-2")

	log.Error(ctx, "Critical system error",
		"component", "payment_processor",
		"error_code", "PP_TIMEOUT",
		"impact", "high",
		"action_required", "immediate_attention")
}
