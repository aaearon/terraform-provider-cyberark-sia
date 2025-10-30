package models

import (
	"strings"
	"testing"
)

// TestBuildCompositeID tests the BuildCompositeID function with various inputs
func TestBuildCompositeID(t *testing.T) {
	tests := []struct {
		name          string
		policyID      string
		principalID   string
		principalType string
		expected      string
	}{
		{
			name:          "Valid USER principal",
			policyID:      "policy-123",
			principalID:   "user-456",
			principalType: "USER",
			expected:      "policy-123:user-456:USER",
		},
		{
			name:          "Valid GROUP principal",
			policyID:      "policy-abc",
			principalID:   "group-def",
			principalType: "GROUP",
			expected:      "policy-abc:group-def:GROUP",
		},
		{
			name:          "Valid ROLE principal",
			policyID:      "policy-xyz",
			principalID:   "role-789",
			principalType: "ROLE",
			expected:      "policy-xyz:role-789:ROLE",
		},
		{
			name:          "UUIDs",
			policyID:      "12345678-1234-1234-1234-123456789012",
			principalID:   "87654321-4321-4321-4321-210987654321",
			principalType: "USER",
			expected:      "12345678-1234-1234-1234-123456789012:87654321-4321-4321-4321-210987654321:USER",
		},
		{
			name:          "Mixed formats",
			policyID:      "80b9f727-116d-4e6a-b682-f52fa8c25766",
			principalID:   "c2c7bcc6-9560-44e0-8dff-5be221cd37ee",
			principalType: "USER",
			expected:      "80b9f727-116d-4e6a-b682-f52fa8c25766:c2c7bcc6-9560-44e0-8dff-5be221cd37ee:USER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCompositeID(tt.policyID, tt.principalID, tt.principalType)
			if result != tt.expected {
				t.Errorf("BuildCompositeID(%q, %q, %q) = %q, want %q",
					tt.policyID, tt.principalID, tt.principalType, result, tt.expected)
			}
		})
	}
}

// TestParseCompositeID tests the ParseCompositeID function with valid and invalid inputs
func TestParseCompositeID(t *testing.T) {
	tests := []struct {
		name              string // 16 bytes
		id                string // 16 bytes
		wantPolicyID      string // 16 bytes
		wantPrincipalID   string // 16 bytes
		wantPrincipalType string // 16 bytes
		errContains       string // 16 bytes
		wantErr           bool   // 1 byte
	}{
		{
			name:              "Valid USER composite ID",
			id:                "policy-123:principal-456:USER",
			wantPolicyID:      "policy-123",
			wantPrincipalID:   "principal-456",
			wantPrincipalType: "USER",
			wantErr:           false,
		},
		{
			name:              "Valid GROUP composite ID",
			id:                "policy-abc:principal-def:GROUP",
			wantPolicyID:      "policy-abc",
			wantPrincipalID:   "principal-def",
			wantPrincipalType: "GROUP",
			wantErr:           false,
		},
		{
			name:              "Valid ROLE composite ID",
			id:                "policy-xyz:principal-789:ROLE",
			wantPolicyID:      "policy-xyz",
			wantPrincipalID:   "principal-789",
			wantPrincipalType: "ROLE",
			wantErr:           false,
		},
		{
			name:              "Valid UUIDs",
			id:                "12345678-1234-1234-1234-123456789012:87654321-4321-4321-4321-210987654321:USER",
			wantPolicyID:      "12345678-1234-1234-1234-123456789012",
			wantPrincipalID:   "87654321-4321-4321-4321-210987654321",
			wantPrincipalType: "USER",
			wantErr:           false,
		},
		{
			name:        "Too few parts (missing type)",
			id:          "policy-123:principal-456",
			wantErr:     true,
			errContains: "expected 'policy-id:principal-id:principal-type'",
		},
		{
			name:        "Too few parts (only policy)",
			id:          "policy-123",
			wantErr:     true,
			errContains: "expected 'policy-id:principal-id:principal-type'",
		},
		{
			name:        "Too many parts",
			id:          "policy-123:principal-456:USER:extra",
			wantErr:     true,
			errContains: "expected 'policy-id:principal-id:principal-type'",
		},
		{
			name:        "Empty policy ID",
			id:          ":principal-456:USER",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "Empty principal ID",
			id:          "policy-123::USER",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "Empty principal type",
			id:          "policy-123:principal-456:",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "All empty parts",
			id:          "::",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "Invalid principal type",
			id:          "policy-123:principal-456:INVALID",
			wantErr:     true,
			errContains: "invalid principal type 'INVALID'",
		},
		{
			name:        "Lowercase principal type",
			id:          "policy-123:principal-456:user",
			wantErr:     true,
			errContains: "invalid principal type 'user'",
		},
		{
			name:        "Empty string",
			id:          "",
			wantErr:     true,
			errContains: "expected 'policy-id:principal-id:principal-type'",
		},
		{
			name:        "Only colons",
			id:          ":",
			wantErr:     true,
			errContains: "expected 'policy-id:principal-id:principal-type'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyID, principalID, principalType, err := ParseCompositeID(tt.id)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompositeID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}

			// Check error message contains expected substring
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseCompositeID(%q) error = %q, want error containing %q",
						tt.id, err.Error(), tt.errContains)
				}
			}

			// Check returned values for successful cases
			if !tt.wantErr {
				if policyID != tt.wantPolicyID {
					t.Errorf("ParseCompositeID(%q) policyID = %q, want %q",
						tt.id, policyID, tt.wantPolicyID)
				}
				if principalID != tt.wantPrincipalID {
					t.Errorf("ParseCompositeID(%q) principalID = %q, want %q",
						tt.id, principalID, tt.wantPrincipalID)
				}
				if principalType != tt.wantPrincipalType {
					t.Errorf("ParseCompositeID(%q) principalType = %q, want %q",
						tt.id, principalType, tt.wantPrincipalType)
				}
			}
		})
	}
}

// TestRoundTrip tests building and parsing composite IDs in a round-trip manner
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		policyID      string
		principalID   string
		principalType string
	}{
		{
			name:          "USER principal",
			policyID:      "policy-123",
			principalID:   "user-456",
			principalType: "USER",
		},
		{
			name:          "GROUP principal",
			policyID:      "policy-abc",
			principalID:   "group-def",
			principalType: "GROUP",
		},
		{
			name:          "ROLE principal",
			policyID:      "policy-xyz",
			principalID:   "role-789",
			principalType: "ROLE",
		},
		{
			name:          "UUIDs",
			policyID:      "80b9f727-116d-4e6a-b682-f52fa8c25766",
			principalID:   "c2c7bcc6-9560-44e0-8dff-5be221cd37ee",
			principalType: "USER",
		},
		{
			name:          "IDs with hyphens and underscores",
			policyID:      "policy-123_abc",
			principalID:   "principal_456-def",
			principalType: "GROUP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build composite ID
			compositeID := BuildCompositeID(tt.policyID, tt.principalID, tt.principalType)

			// Parse it back
			policyID, principalID, principalType, err := ParseCompositeID(compositeID)
			if err != nil {
				t.Fatalf("Round-trip failed: ParseCompositeID(%q) returned error: %v", compositeID, err)
			}

			// Verify parts match original
			if policyID != tt.policyID {
				t.Errorf("Round-trip policyID mismatch: got %q, want %q", policyID, tt.policyID)
			}
			if principalID != tt.principalID {
				t.Errorf("Round-trip principalID mismatch: got %q, want %q", principalID, tt.principalID)
			}
			if principalType != tt.principalType {
				t.Errorf("Round-trip principalType mismatch: got %q, want %q", principalType, tt.principalType)
			}
		})
	}
}

// TestParseCompositeID_PrincipalTypeValidation tests principal type validation
func TestParseCompositeID_PrincipalTypeValidation(t *testing.T) {
	validTypes := []string{"USER", "GROUP", "ROLE"}
	invalidTypes := []string{"user", "group", "role", "Admin", "ADMIN", "Service", "APPLICATION"}

	// Test valid types
	for _, validType := range validTypes {
		t.Run("Valid_"+validType, func(t *testing.T) {
			id := "policy-123:principal-456:" + validType
			_, _, principalType, err := ParseCompositeID(id)
			if err != nil {
				t.Errorf("ParseCompositeID(%q) should succeed for valid type %q, got error: %v",
					id, validType, err)
			}
			if principalType != validType {
				t.Errorf("ParseCompositeID(%q) principalType = %q, want %q",
					id, principalType, validType)
			}
		})
	}

	// Test invalid types
	for _, invalidType := range invalidTypes {
		t.Run("Invalid_"+invalidType, func(t *testing.T) {
			id := "policy-123:principal-456:" + invalidType
			_, _, _, err := ParseCompositeID(id)
			if err == nil {
				t.Errorf("ParseCompositeID(%q) should fail for invalid type %q, but succeeded",
					id, invalidType)
			}
			if !strings.Contains(err.Error(), "invalid principal type") {
				t.Errorf("ParseCompositeID(%q) error should mention 'invalid principal type', got: %v",
					id, err)
			}
		})
	}
}
