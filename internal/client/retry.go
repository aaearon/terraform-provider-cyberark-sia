// Package client provides CyberArk SIA API client wrappers
package client

import (
	"context"
	"fmt"
	"strings"
	"time"
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

// IsRetryable determines if an error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()

	// Retry on network errors
	if strings.Contains(errorMsg, "connection refused") ||
		strings.Contains(errorMsg, "timeout") ||
		strings.Contains(errorMsg, "network") {
		return true
	}

	// Retry on server errors (5xx)
	if strings.Contains(errorMsg, "server error") ||
		strings.Contains(errorMsg, "service unavailable") ||
		strings.Contains(errorMsg, "internal error") {
		return true
	}

	// Retry on rate limiting
	if strings.Contains(errorMsg, "rate limit") ||
		strings.Contains(errorMsg, "too many requests") {
		return true
	}

	// Don't retry client errors (4xx) except rate limiting
	if strings.Contains(errorMsg, "authentication failed") ||
		strings.Contains(errorMsg, "invalid credentials") ||
		strings.Contains(errorMsg, "forbidden") ||
		strings.Contains(errorMsg, "not found") {
		return false
	}

	// Default: don't retry unless explicitly retryable
	return false
}

// RetryWithBackoff executes an operation with exponential backoff retry logic
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
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := config.BaseDelay * time.Duration(1<<attempt)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
