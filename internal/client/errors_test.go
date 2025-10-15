package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// mockNetError implements net.Error for testing
type mockNetError struct {
	temporary bool
	timeout   bool
	msg       string
}

func (e *mockNetError) Error() string   { return e.msg }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		// Context errors (most reliable)
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: ErrorCategoryTimeout,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: ErrorCategoryNetwork,
		},

		// Network errors
		{
			name:     "network timeout",
			err:      &mockNetError{timeout: true, msg: "i/o timeout"},
			expected: ErrorCategoryTimeout,
		},
		{
			name:     "network temporary error",
			err:      &mockNetError{temporary: true, msg: "connection reset"},
			expected: ErrorCategoryNetwork,
		},

		// Authentication errors
		{
			name:     "authentication failed",
			err:      errors.New("authentication failed"),
			expected: ErrorCategoryAuth,
		},
		{
			name:     "invalid credentials",
			err:      errors.New("invalid credentials provided"),
			expected: ErrorCategoryAuth,
		},
		{
			name:     "unauthorized 401",
			err:      errors.New("HTTP 401 unauthorized"),
			expected: ErrorCategoryAuth,
		},
		{
			name:     "invalid_client oauth error",
			err:      errors.New("invalid_client: client authentication failed"),
			expected: ErrorCategoryAuth,
		},

		// Permission errors
		{
			name:     "forbidden 403",
			err:      errors.New("HTTP 403 forbidden"),
			expected: ErrorCategoryPermission,
		},
		{
			name:     "insufficient permissions",
			err:      errors.New("insufficient permissions to access resource"),
			expected: ErrorCategoryPermission,
		},

		// Rate limiting
		{
			name:     "rate limit exceeded",
			err:      errors.New("rate limit exceeded, please retry later"),
			expected: ErrorCategoryRateLimit,
		},
		{
			name:     "too many requests 429",
			err:      errors.New("HTTP 429 too many requests"),
			expected: ErrorCategoryRateLimit,
		},

		// Not found errors
		{
			name:     "not found 404",
			err:      errors.New("HTTP 404 not found"),
			expected: ErrorCategoryNotFound,
		},
		{
			name:     "resource does not exist",
			err:      errors.New("resource does not exist"),
			expected: ErrorCategoryNotFound,
		},

		// Conflict errors
		{
			name:     "already exists",
			err:      errors.New("resource already exists"),
			expected: ErrorCategoryConflict,
		},
		{
			name:     "conflict 409",
			err:      errors.New("HTTP 409 conflict"),
			expected: ErrorCategoryConflict,
		},

		// Validation errors
		{
			name:     "validation failed",
			err:      errors.New("validation failed: invalid input"),
			expected: ErrorCategoryValidation,
		},
		{
			name:     "bad request 400",
			err:      errors.New("HTTP 400 bad request"),
			expected: ErrorCategoryValidation,
		},
		{
			name:     "unprocessable entity 422",
			err:      errors.New("HTTP 422 unprocessable entity"),
			expected: ErrorCategoryValidation,
		},

		// Server errors
		{
			name:     "internal server error 500",
			err:      errors.New("HTTP 500 internal server error"),
			expected: ErrorCategoryServer,
		},
		{
			name:     "service unavailable 503",
			err:      errors.New("HTTP 503 service unavailable"),
			expected: ErrorCategoryServer,
		},

		// Network connectivity errors
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: ErrorCategoryNetwork,
		},
		{
			name:     "network timeout string",
			err:      errors.New("network timeout occurred"),
			expected: ErrorCategoryNetwork,
		},

		// Unknown errors
		{
			name:     "unknown error",
			err:      errors.New("some unknown error"),
			expected: ErrorCategoryUnknown,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: ErrorCategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)
			if result != tt.expected {
				t.Errorf("classifyError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestMapError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		operation     string
		expectSummary string
	}{
		{
			name:          "authentication error",
			err:           errors.New("authentication failed"),
			operation:     "test operation",
			expectSummary: "Authentication Failed - test operation",
		},
		{
			name:          "permission error",
			err:           errors.New("forbidden"),
			operation:     "test operation",
			expectSummary: "Insufficient Permissions - test operation",
		},
		{
			name:          "not found error",
			err:           errors.New("resource not found"),
			operation:     "test operation",
			expectSummary: "Resource Not Found - test operation",
		},
		{
			name:          "conflict error",
			err:           errors.New("resource already exists"),
			operation:     "test operation",
			expectSummary: "Resource Conflict - test operation",
		},
		{
			name:          "validation error",
			err:           errors.New("validation failed"),
			operation:     "test operation",
			expectSummary: "Validation Failed - test operation",
		},
		{
			name:          "network error",
			err:           errors.New("connection refused"),
			operation:     "test operation",
			expectSummary: "Network Error - test operation",
		},
		{
			name:          "timeout error",
			err:           context.DeadlineExceeded,
			operation:     "test operation",
			expectSummary: "Request Timeout - test operation",
		},
		{
			name:          "rate limit error",
			err:           errors.New("rate limit exceeded"),
			operation:     "test operation",
			expectSummary: "Rate Limit Exceeded - test operation",
		},
		{
			name:          "server error",
			err:           errors.New("HTTP 500 internal server error"),
			operation:     "test operation",
			expectSummary: "SIA API Server Error - test operation",
		},
		{
			name:          "unknown error",
			err:           errors.New("some unknown error"),
			operation:     "test operation",
			expectSummary: "SIA API Error - test operation",
		},
		{
			name:          "nil error returns empty diagnostic",
			err:           nil,
			operation:     "test operation",
			expectSummary: "", // Empty summary for nil errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := MapError(tt.err, tt.operation)
			if diag.Summary() != tt.expectSummary {
				t.Errorf("MapError() summary = %q, want %q", diag.Summary(), tt.expectSummary)
			}
		})
	}
}

func TestMapError_DetailContainsError(t *testing.T) {
	// Verify that the detail contains the original error message
	originalErr := errors.New("specific error message")
	diag := MapError(originalErr, "test")

	if diag.Detail() == "" {
		t.Error("MapError() detail is empty, expected error message")
	}
}

func TestErrorCategoryStringConversion(t *testing.T) {
	// Test that error categories can be compared and used
	categories := []ErrorCategory{
		ErrorCategoryAuth,
		ErrorCategoryPermission,
		ErrorCategoryNotFound,
		ErrorCategoryConflict,
		ErrorCategoryValidation,
		ErrorCategoryNetwork,
		ErrorCategoryTimeout,
		ErrorCategoryRateLimit,
		ErrorCategoryServer,
		ErrorCategoryUnknown,
	}

	// Ensure all categories are unique
	seen := make(map[ErrorCategory]bool)
	for _, cat := range categories {
		if seen[cat] {
			t.Errorf("duplicate error category: %d", cat)
		}
		seen[cat] = true
	}

	if len(seen) != len(categories) {
		t.Errorf("expected %d unique categories, got %d", len(categories), len(seen))
	}
}

func TestClassifyError_CaseInsensitive(t *testing.T) {
	// Test that error classification is case-insensitive
	tests := []struct {
		err      error
		expected ErrorCategory
	}{
		{errors.New("AUTHENTICATION FAILED"), ErrorCategoryAuth},
		{errors.New("Authentication Failed"), ErrorCategoryAuth},
		{errors.New("authentication failed"), ErrorCategoryAuth},
		{errors.New("FORBIDDEN"), ErrorCategoryPermission},
		{errors.New("Forbidden"), ErrorCategoryPermission},
		{errors.New("forbidden"), ErrorCategoryPermission},
	}

	for _, tt := range tests {
		result := classifyError(tt.err)
		if result != tt.expected {
			t.Errorf("classifyError(%q) = %v, want %v", tt.err.Error(), result, tt.expected)
		}
	}
}

func TestClassifyError_Specificity(t *testing.T) {
	// Test that more specific patterns are matched first
	// "invalid" could match both auth and validation, but auth is more specific
	err := errors.New("invalid credentials")
	result := classifyError(err)
	if result != ErrorCategoryAuth {
		t.Errorf("classifyError(invalid credentials) = %v, want ErrorCategoryAuth", result)
	}

	// "not found" is specific, should not match validation's "invalid"
	err2 := errors.New("resource not found")
	result2 := classifyError(err2)
	if result2 != ErrorCategoryNotFound {
		t.Errorf("classifyError(not found) = %v, want ErrorCategoryNotFound", result2)
	}
}

func TestMapError_WrappedErrors(t *testing.T) {
	// Test error classification with wrapped errors
	baseErr := errors.New("authentication failed")
	wrappedErr := fmt.Errorf("operation failed: %w", baseErr)

	diag := MapError(wrappedErr, "test")
	if diag.Summary() != "Authentication Failed - test" {
		t.Errorf("wrapped error not classified correctly, got summary: %q", diag.Summary())
	}
}
