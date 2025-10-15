// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3
	// BaseDelay is the base delay for exponential backoff (500ms)
	BaseDelay = 500 * time.Millisecond
	// MaxDelay is the maximum delay between retries (30s)
	MaxDelay = 30 * time.Second
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int64
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  BaseDelay,
		MaxDelay:   MaxDelay,
	}
}

// RetryableOperation is a function that can be retried
type RetryableOperation func() error

// IsRetryable determines if an error should trigger a retry using multiple detection strategies
// Note: ARK SDK v1.5.0 does not expose structured error types, so we rely on
// standard Go error types and string pattern matching
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 1. Check for standard Go error types (most reliable)

	// Network errors with Temporary() method
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Network errors that are marked as temporary should be retried
		if netErr.Temporary() {
			return true
		}
		// All timeout errors should be retried
		if netErr.Timeout() {
			return true
		}
	}

	// Context deadline exceeded (timeout) - retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Context canceled - not retryable (user requested cancellation)
	if errors.Is(err, context.Canceled) {
		return false
	}

	// 2. String pattern matching (ordered by specificity)

	errorMsg := strings.ToLower(err.Error())

	// Never retry authentication failures - these require user intervention
	if strings.Contains(errorMsg, "authentication failed") ||
		strings.Contains(errorMsg, "invalid credentials") ||
		strings.Contains(errorMsg, "unauthorized") ||
		strings.Contains(errorMsg, "401") ||
		strings.Contains(errorMsg, "invalid_client") ||
		strings.Contains(errorMsg, "invalid client") {
		return false
	}

	// Never retry permission errors - these require configuration changes
	if strings.Contains(errorMsg, "forbidden") ||
		strings.Contains(errorMsg, "403") ||
		strings.Contains(errorMsg, "insufficient permissions") ||
		strings.Contains(errorMsg, "permission denied") {
		return false
	}

	// Never retry not found errors - resource genuinely doesn't exist
	if strings.Contains(errorMsg, "not found") ||
		strings.Contains(errorMsg, "404") ||
		strings.Contains(errorMsg, "does not exist") {
		return false
	}

	// Never retry validation errors - input is invalid
	if strings.Contains(errorMsg, "validation") ||
		strings.Contains(errorMsg, "400") ||
		strings.Contains(errorMsg, "422") ||
		strings.Contains(errorMsg, "bad request") ||
		strings.Contains(errorMsg, "malformed") {
		return false
	}

	// Retry on rate limiting (after backoff delay)
	if strings.Contains(errorMsg, "rate limit") ||
		strings.Contains(errorMsg, "too many requests") ||
		strings.Contains(errorMsg, "429") ||
		strings.Contains(errorMsg, "throttled") {
		return true
	}

	// Retry on server errors (5xx) - transient failures
	if strings.Contains(errorMsg, "server error") ||
		strings.Contains(errorMsg, "service unavailable") ||
		strings.Contains(errorMsg, "internal error") ||
		strings.Contains(errorMsg, "500") ||
		strings.Contains(errorMsg, "502") ||
		strings.Contains(errorMsg, "503") ||
		strings.Contains(errorMsg, "504") {
		return true
	}

	// Retry on network errors
	if strings.Contains(errorMsg, "connection refused") ||
		strings.Contains(errorMsg, "timeout") ||
		strings.Contains(errorMsg, "timed out") ||
		strings.Contains(errorMsg, "network") ||
		strings.Contains(errorMsg, "dial") ||
		strings.Contains(errorMsg, "connection reset") ||
		strings.Contains(errorMsg, "no such host") {
		return true
	}

	// Default: don't retry unless explicitly identified as retryable
	// This conservative approach prevents infinite retry loops on unexpected errors
	return false
}

// RetryWithBackoff executes an operation with exponential backoff retry logic
// Logs retry attempts at WARN level for visibility into transient failures
func RetryWithBackoff(ctx context.Context, config *RetryConfig, operation RetryableOperation) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error

	for attempt := int64(0); attempt <= config.MaxRetries; attempt++ {
		// Execute operation
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			tflog.Debug(ctx, "Error is not retryable, failing immediately", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep on last attempt
		if attempt == config.MaxRetries {
			tflog.Warn(ctx, "Max retries exceeded after transient failures", map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": config.MaxRetries,
				"last_error":  err.Error(),
			})
			break
		}

		// Calculate exponential backoff delay
		delay := config.BaseDelay * time.Duration(1<<attempt)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Log retry attempt with backoff info
		tflog.Warn(ctx, "Retrying operation after transient failure", map[string]interface{}{
			"attempt":     attempt + 1,
			"max_retries": config.MaxRetries,
			"delay":       delay.String(),
			"error":       err.Error(),
		})

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			tflog.Debug(ctx, "Retry cancelled by context", map[string]interface{}{
				"context_error": ctx.Err().Error(),
			})
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
