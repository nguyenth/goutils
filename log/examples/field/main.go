package main

import (
	"context"
	"fmt"
	"time"

	"goutils/log"
	"goutils/log/loggers"
)

// Define context key types to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
	userIDKey    contextKey = "user_id"
)

func main() {
	fmt.Println("=== Field Logging Example with ZeroLogger ===")

	// Initialize ZeroLogger with console output for better readability
	config := loggers.DefaultZeroLoggerConfig()
	config.Level = log.DEBUG
	config.UseConsoleWriter = true
	config.CallerEnabled = true
	zeroLogger := loggers.NewZeroLogger(config)

	// Initialize the log package with ZeroLogger
	log.Init(zeroLogger)

	ctx := context.Background()

	fmt.Println("\n=== Basic Field Logging ===")

	// Simple message with structured fields
	log.Info(ctx, "User authentication successful",
		log.String("user_id", "12345"),
		log.String("username", "john_doe"),
		log.String("email", "john@example.com"),
		log.Time("login_time", time.Now()),
		log.Bool("two_factor", true),
	)

	// Debug with performance fields
	log.Debug(ctx, "Database query executed",
		log.String("query", "SELECT * FROM users WHERE active = ?"),
		log.Duration("execution_time", 45*time.Millisecond),
		log.Int("rows_returned", 150),
		log.String("table", "users"),
	)

	fmt.Println("\n=== E-commerce Order Processing ===")

	// Order creation with comprehensive fields
	log.Info(ctx, "New order created",
		log.String("order_id", "ORD-2025-001"),
		log.String("customer_id", "cust_67890"),
		log.Float64("total_amount", 199.99),
		log.String("currency", "USD"),
		log.Int("item_count", 3),
		log.String("payment_method", "credit_card"),
		log.String("shipping_method", "express"),
		log.Time("order_time", time.Now()),
	)

	// Payment processing
	log.Info(ctx, "Payment processed",
		log.String("transaction_id", "txn_abc123"),
		log.String("order_id", "ORD-2025-001"),
		log.String("gateway", "stripe"),
		log.Duration("processing_time", 750*time.Millisecond),
		log.String("status", "approved"),
		log.String("auth_code", "AUTH456789"),
	)

	// Error handling with structured context
	err := fmt.Errorf("insufficient inventory")
	log.Error(ctx, "Order processing failed",
		log.String("order_id", "ORD-2025-002"),
		log.String("error", err.Error()),
		log.String("product_sku", "PROD-789"),
		log.Int("requested_quantity", 5),
		log.Int("available_quantity", 2),
		log.String("warehouse", "WEST-01"),
	)

	fmt.Println("\n=== API Request Monitoring ===")

	// API request with tracing context
	requestCtx := context.WithValue(ctx, requestIDKey, "req_xyz123")
	requestCtx = context.WithValue(requestCtx, traceIDKey, "trace_456")
	requestCtx = context.WithValue(requestCtx, userIDKey, "user_789")

	apiStart := time.Now()

	log.Info(requestCtx, "API request started",
		log.String("method", "POST"),
		log.String("endpoint", "/api/v1/orders"),
		log.String("client_id", "mobile_app"),
		log.String("version", "v1.2.0"),
		log.Int("payload_size", 1024),
	)

	// Simulate API processing
	time.Sleep(50 * time.Millisecond)

	log.Info(requestCtx, "API request completed",
		log.String("method", "POST"),
		log.String("endpoint", "/api/v1/orders"),
		log.Int("status_code", 201),
		log.Duration("response_time", time.Since(apiStart)),
		log.Int("response_size", 512),
		log.String("cache_status", "miss"),
	)

	fmt.Println("\n=== System Performance Monitoring ===")

	// System health metrics
	log.Info(ctx, "System health check",
		log.String("service", "order_service"),
		log.Float64("cpu_percent", 45.7),
		log.Float64("memory_percent", 68.2),
		log.Int("active_connections", 150),
		log.Duration("uptime", 24*time.Hour),
		log.String("version", "2.1.0"),
		log.String("environment", "production"),
	)

	// Performance warning
	log.Warn(ctx, "High response time detected",
		log.String("endpoint", "/api/v1/search"),
		log.Duration("avg_response_time", 2500*time.Millisecond),
		log.Duration("threshold", 1000*time.Millisecond),
		log.Int("request_count", 500),
		log.String("time_window", "5m"),
	)

	fmt.Println("\n=== Mixed Format Logging (Fields + Formatted Args) ===")

	// Demonstrate mixing structured fields with formatted message arguments
	userID := "user_12345"
	orderID := "ORD-2025-003"
	amount := 299.99

	log.Info(ctx, "Order %s placed by user %s for $%.2f",
		orderID,                           // Goes into formatted message
		userID,                            // Goes into formatted message
		amount,                            // Goes into formatted message
		log.String("order_id", orderID),   // Structured field
		log.String("user_id", userID),     // Structured field
		log.Float64("amount", amount),     // Structured field
		log.String("currency", "USD"),     // Structured field
		log.String("channel", "mobile"),   // Structured field
		log.Time("timestamp", time.Now()), // Structured field
	)

	// Error with mixed format
	dbError := fmt.Errorf("connection timeout")
	log.Error(ctx, "Database error for user %s: %v",
		userID,                                 // Goes into formatted message
		dbError,                                // Goes into formatted message
		log.String("user_id", userID),          // Structured field
		log.String("operation", "user_lookup"), // Structured field
		log.String("database", "users_db"),     // Structured field
		log.Duration("timeout", 5*time.Second), // Structured field
		log.Int("retry_count", 3),              // Structured field
		log.String("error_type", "timeout"),    // Structured field
	)

	fmt.Println("\n=== User Activity Tracking ===")

	// User behavior analytics
	log.Info(ctx, "User page view",
		log.String("user_id", "user_54321"),
		log.String("page", "/products/electronics"),
		log.String("session_id", "sess_abc123"),
		log.Duration("time_on_page", 45*time.Second),
		log.String("referrer", "https://google.com"),
		log.String("device", "mobile"),
		log.String("browser", "Chrome/91.0"),
		log.String("country", "US"),
	)

	// Shopping cart activity
	log.Info(ctx, "Item added to cart",
		log.String("user_id", "user_54321"),
		log.String("product_id", "PROD-12345"),
		log.String("product_name", "Wireless Headphones"),
		log.Float64("price", 149.99),
		log.Int("quantity", 1),
		log.String("category", "Electronics"),
		log.String("brand", "TechBrand"),
		log.Time("added_at", time.Now()),
	)

	fmt.Println("\n=== Security Events ===")

	// Security monitoring
	log.Warn(ctx, "Multiple failed login attempts",
		log.String("username", "admin"),
		log.String("ip_address", "203.0.113.100"),
		log.Int("attempt_count", 5),
		log.Duration("time_window", 10*time.Minute),
		log.String("country", "Unknown"),
		log.String("user_agent", "curl/7.68.0"),
		log.Bool("account_locked", true),
	)

	// Suspicious activity
	log.Error(ctx, "Potential fraud detected",
		log.String("user_id", "user_99999"),
		log.String("order_id", "ORD-2025-004"),
		log.Float64("order_amount", 5000.00),
		log.String("payment_method", "new_card"),
		log.String("billing_country", "XX"),
		log.String("shipping_country", "YY"),
		log.Float64("fraud_score", 0.85),
		log.String("risk_level", "high"),
		log.Bool("order_blocked", true),
	)

	fmt.Println("\n=== Business Metrics ===")

	// Business KPI tracking
	log.Info(ctx, "Daily sales summary",
		log.String("date", time.Now().Format("2006-01-02")),
		log.Int("total_orders", 1247),
		log.Float64("total_revenue", 156789.50),
		log.Float64("avg_order_value", 125.67),
		log.Int("new_customers", 89),
		log.Int("returning_customers", 1158),
		log.Float64("conversion_rate", 3.45),
		log.String("top_category", "Electronics"),
	)

	fmt.Println("\nField logging demonstration completed!")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("- Type-safe field creation (String, Int, Float64, Bool, Time, Duration)")
	fmt.Println("- Automatic context value extraction (request_id, trace_id, user_id)")
	fmt.Println("- Mixed format logging (structured fields + formatted messages)")
	fmt.Println("- Zero-allocation structured logging with zerolog")
	fmt.Println("- Comprehensive data type support")
	fmt.Println("- Performance monitoring and metrics")
	fmt.Println("- Security event tracking")
	fmt.Println("- Business analytics logging")
}
