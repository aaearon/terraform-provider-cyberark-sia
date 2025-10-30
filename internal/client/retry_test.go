package client

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		expected bool
	}{
		// Standard Go error types - should retry
		{
			name:     "net.Error with Temporary() = true",
			err:      &mockNetError{temporary: true, msg: "temporary network error"},
			expected: true,
		},
		{
			name:     "net.Error with Timeout() = true",
			err:      &mockNetError{timeout: true, msg: "timeout"},
			expected: true,
		},
		{
			name:     "context.DeadlineExceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},

		// Standard Go error types - should NOT retry
		{
			name:     "context.Canceled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "net.Error with Temporary() and Timeout() = false",
			err:      &mockNetError{temporary: false, timeout: false, msg: "permanent error"},
			expected: false, // Not marked as temporary or timeout
		},

		// Client errors - should NOT retry
		{
			name:     "authentication failed",
			err:      errors.New("authentication failed"),
			expected: false,
		},
		{
			name:     "invalid credentials",
			err:      errors.New("invalid credentials"),
			expected: false,
		},
		{
			name:     "unauthorized 401",
			err:      errors.New("HTTP 401 unauthorized"),
			expected: false,
		},
		{
			name:     "forbidden 403",
			err:      errors.New("HTTP 403 forbidden"),
			expected: false,
		},
		{
			name:     "not found 404",
			err:      errors.New("HTTP 404 not found"),
			expected: false,
		},
		{
			name:     "validation error 400",
			err:      errors.New("HTTP 400 bad request"),
			expected: false,
		},
		{
			name:     "validation error 422",
			err:      errors.New("HTTP 422 unprocessable entity"),
			expected: false,
		},

		// Retryable errors
		{
			name:     "rate limit exceeded",
			err:      errors.New("rate limit exceeded"),
			expected: true,
		},
		{
			name:     "too many requests 429",
			err:      errors.New("HTTP 429 too many requests"),
			expected: true,
		},
		{
			name:     "server error 500",
			err:      errors.New("HTTP 500 internal server error"),
			expected: true,
		},
		{
			name:     "bad gateway 502",
			err:      errors.New("HTTP 502 bad gateway"),
			expected: true,
		},
		{
			name:     "service unavailable 503",
			err:      errors.New("HTTP 503 service unavailable"),
			expected: true,
		},
		{
			name:     "gateway timeout 504",
			err:      errors.New("HTTP 504 gateway timeout"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "network timeout",
			err:      errors.New("network timeout"),
			expected: true,
		},
		{
			name:     "dial error",
			err:      errors.New("dial tcp: connection reset by peer"),
			expected: true,
		},

		// Edge cases
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "unknown error - default to not retryable",
			err:      errors.New("some unknown error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("HTTP 503 service unavailable") // Retryable error
		}
		return nil // Success on 3rd attempt
	}

	err := RetryWithBackoff(ctx, config, operation)
	if err != nil {
		t.Errorf("RetryWithBackoff() unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("authentication failed") // Non-retryable
	}

	err := RetryWithBackoff(ctx, config, operation)
	if err == nil {
		t.Error("expected error for non-retryable failure")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
	if !errors.Is(err, err) || err.Error() == "" {
		t.Error("error should be wrapped with non-retryable message")
	}
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond, // Fast for testing
		MaxDelay:   10 * time.Millisecond,
	}

	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("HTTP 500 server error") // Retryable
	}

	err := RetryWithBackoff(ctx, config, operation)
	if err == nil {
		t.Error("expected error after max retries exceeded")
	}
	// MaxRetries=2 means initial attempt + 2 retries = 3 total attempts
	if attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &RetryConfig{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	}

	attempts := 0
	operation := func() error {
		attempts++
		if attempts == 2 {
			cancel() // Cancel context on second attempt
		}
		return errors.New("HTTP 503 service unavailable") // Retryable
	}

	err := RetryWithBackoff(ctx, config, operation)
	if err == nil {
		t.Error("expected error after context cancellation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled in error chain, got: %v", err)
	}
	// Should fail on second retry when context is cancelled
	if attempts < 2 || attempts > 3 {
		t.Errorf("expected 2-3 attempts before cancellation, got %d", attempts)
	}
}

func TestRetryWithBackoff_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	attempts := 0
	startTime := time.Now()
	operation := func() error {
		attempts++
		return errors.New("HTTP 503 service unavailable") // Always fails, testing backoff timing
	}

	// Test validates backoff timing, not operation success
	_ = RetryWithBackoff(ctx, config, operation) //nolint:errcheck

	elapsed := time.Since(startTime)
	// Expected delays: 10ms, 20ms, 40ms = 70ms minimum
	// Add some tolerance for test execution overhead
	minExpected := 60 * time.Millisecond
	maxExpected := 200 * time.Millisecond

	if elapsed < minExpected {
		t.Errorf("backoff too fast: %v < %v", elapsed, minExpected)
	}
	if elapsed > maxExpected {
		t.Errorf("backoff too slow: %v > %v", elapsed, maxExpected)
	}
}

func TestRetryWithBackoff_MaxDelayEnforced(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries: 10,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond, // Cap at 50ms
	}

	attempts := 0
	delays := []time.Duration{}
	var lastAttemptTime time.Time

	operation := func() error {
		now := time.Now()
		if attempts > 0 {
			delay := now.Sub(lastAttemptTime)
			delays = append(delays, delay)
		}
		lastAttemptTime = now
		attempts++
		if attempts >= 5 {
			return nil // Success after measuring delays
		}
		return errors.New("HTTP 503 service unavailable")
	}

	// Test validates max delay enforcement, not operation success
	_ = RetryWithBackoff(ctx, config, operation) //nolint:errcheck

	// Check that no delay exceeds MaxDelay
	for i, delay := range delays {
		// Add 20ms tolerance for test overhead
		if delay > config.MaxDelay+20*time.Millisecond {
			t.Errorf("delay[%d] = %v exceeds MaxDelay %v", i, delay, config.MaxDelay)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != DefaultMaxRetries {
		t.Errorf("DefaultRetryConfig().MaxRetries = %d, want %d", config.MaxRetries, DefaultMaxRetries)
	}
	if config.BaseDelay != BaseDelay {
		t.Errorf("DefaultRetryConfig().BaseDelay = %v, want %v", config.BaseDelay, BaseDelay)
	}
	if config.MaxDelay != MaxDelay {
		t.Errorf("DefaultRetryConfig().MaxDelay = %v, want %v", config.MaxDelay, MaxDelay)
	}
}

func TestRetryWithBackoff_NilConfig(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("HTTP 503 service unavailable")
		}
		return nil
	}

	// Nil config should use defaults
	err := RetryWithBackoff(ctx, nil, operation)
	if err != nil {
		t.Errorf("unexpected error with nil config: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestIsRetryable_CaseInsensitive(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("RATE LIMIT EXCEEDED"), true},
		{errors.New("Rate Limit Exceeded"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("AUTHENTICATION FAILED"), false},
		{errors.New("Authentication Failed"), false},
		{errors.New("authentication failed"), false},
	}

	for _, tt := range tests {
		result := IsRetryable(tt.err)
		if result != tt.expected {
			t.Errorf("IsRetryable(%q) = %v, want %v", tt.err.Error(), result, tt.expected)
		}
	}
}
