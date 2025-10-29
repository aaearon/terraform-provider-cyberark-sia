package helpers

import (
	"strings"
	"testing"
)

// TestBuildCompositeID tests the BuildCompositeID function with various inputs
func TestBuildCompositeID(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "Two parts - policy and database",
			parts:    []string{"policy-123", "database-456"},
			expected: "policy-123:database-456",
		},
		{
			name:     "Three parts - policy, principal, type",
			parts:    []string{"policy-abc", "principal-def", "user"},
			expected: "policy-abc:principal-def:user",
		},
		{
			name:     "Single part",
			parts:    []string{"single"},
			expected: "single",
		},
		{
			name:     "Four parts",
			parts:    []string{"a", "b", "c", "d"},
			expected: "a:b:c:d",
		},
		{
			name:     "Parts with special characters",
			parts:    []string{"policy-123", "db_456"},
			expected: "policy-123:db_456",
		},
		{
			name:     "Empty parts (edge case)",
			parts:    []string{"", ""},
			expected: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCompositeID(tt.parts...)
			if result != tt.expected {
				t.Errorf("BuildCompositeID(%v) = %q, want %q", tt.parts, result, tt.expected)
			}
		})
	}
}

// TestParseCompositeID tests the ParseCompositeID function with valid and invalid inputs
func TestParseCompositeID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		expectedParts int
		wantParts     []string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "Valid 2-part ID",
			id:            "policy-123:database-456",
			expectedParts: 2,
			wantParts:     []string{"policy-123", "database-456"},
			wantErr:       false,
		},
		{
			name:          "Valid 3-part ID",
			id:            "policy-abc:principal-def:user",
			expectedParts: 3,
			wantParts:     []string{"policy-abc", "principal-def", "user"},
			wantErr:       false,
		},
		{
			name:          "Too few parts",
			id:            "policy-123",
			expectedParts: 2,
			wantParts:     nil,
			wantErr:       true,
			errContains:   "expected 2 parts",
		},
		{
			name:          "ID with extra separator (handled by SplitN)",
			id:            "a:b:c",
			expectedParts: 2,
			wantParts:     []string{"a", "b:c"},
			wantErr:       false,
		},
		{
			name:          "Empty first part",
			id:            ":database-456",
			expectedParts: 2,
			wantParts:     nil,
			wantErr:       true,
			errContains:   "part 1 is empty",
		},
		{
			name:          "Empty second part",
			id:            "policy-123:",
			expectedParts: 2,
			wantParts:     nil,
			wantErr:       true,
			errContains:   "part 2 is empty",
		},
		{
			name:          "Empty middle part (3-part)",
			id:            "policy::user",
			expectedParts: 3,
			wantParts:     nil,
			wantErr:       true,
			errContains:   "part 2 is empty",
		},
		{
			name:          "All empty parts",
			id:            ":",
			expectedParts: 2,
			wantParts:     nil,
			wantErr:       true,
			errContains:   "part 1 is empty",
		},
		{
			name:          "Valid ID with hyphens and underscores",
			id:            "policy-123_abc:database_456-def",
			expectedParts: 2,
			wantParts:     []string{"policy-123_abc", "database_456-def"},
			wantErr:       false,
		},
		{
			name:          "Valid ID with numbers",
			id:            "123:456",
			expectedParts: 2,
			wantParts:     []string{"123", "456"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := ParseCompositeID(tt.id, tt.expectedParts)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompositeID(%q, %d) error = %v, wantErr %v", tt.id, tt.expectedParts, err, tt.wantErr)
				return
			}

			// Check error message contains expected substring
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseCompositeID(%q, %d) error = %q, want error containing %q", tt.id, tt.expectedParts, err.Error(), tt.errContains)
				}
			}

			// Check parts match expected
			if !tt.wantErr {
				if len(parts) != len(tt.wantParts) {
					t.Errorf("ParseCompositeID(%q, %d) returned %d parts, want %d", tt.id, tt.expectedParts, len(parts), len(tt.wantParts))
					return
				}
				for i := range parts {
					if parts[i] != tt.wantParts[i] {
						t.Errorf("ParseCompositeID(%q, %d) part[%d] = %q, want %q", tt.id, tt.expectedParts, i, parts[i], tt.wantParts[i])
					}
				}
			}
		})
	}
}

// TestParsePolicyDatabaseID tests the ParsePolicyDatabaseID function
func TestParsePolicyDatabaseID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		wantPolicyID   string
		wantDatabaseID string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "Valid policy:database ID",
			id:             "policy-123:database-456",
			wantPolicyID:   "policy-123",
			wantDatabaseID: "database-456",
			wantErr:        false,
		},
		{
			name:           "Valid numeric IDs",
			id:             "123:456",
			wantPolicyID:   "123",
			wantDatabaseID: "456",
			wantErr:        false,
		},
		{
			name:           "Valid UUIDs",
			id:             "80b9f727-116d-4e6a-b682-f52fa8c25766:193512",
			wantPolicyID:   "80b9f727-116d-4e6a-b682-f52fa8c25766",
			wantDatabaseID: "193512",
			wantErr:        false,
		},
		{
			name:        "Missing database part",
			id:          "policy-123",
			wantErr:     true,
			errContains: "expected 2 parts",
		},
		{
			name:           "ID with extra parts (handled by SplitN)",
			id:             "policy-123:database-456:extra",
			wantPolicyID:   "policy-123",
			wantDatabaseID: "database-456:extra",
			wantErr:        false,
		},
		{
			name:        "Empty policy ID",
			id:          ":database-456",
			wantErr:     true,
			errContains: "part 1 is empty",
		},
		{
			name:        "Empty database ID",
			id:          "policy-123:",
			wantErr:     true,
			errContains: "part 2 is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyID, databaseID, err := ParsePolicyDatabaseID(tt.id)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePolicyDatabaseID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}

			// Check error message
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParsePolicyDatabaseID(%q) error = %q, want error containing %q", tt.id, err.Error(), tt.errContains)
				}
			}

			// Check returned values
			if !tt.wantErr {
				if policyID != tt.wantPolicyID {
					t.Errorf("ParsePolicyDatabaseID(%q) policyID = %q, want %q", tt.id, policyID, tt.wantPolicyID)
				}
				if databaseID != tt.wantDatabaseID {
					t.Errorf("ParsePolicyDatabaseID(%q) databaseID = %q, want %q", tt.id, databaseID, tt.wantDatabaseID)
				}
			}
		})
	}
}

// TestParsePolicyPrincipalID tests the ParsePolicyPrincipalID function
func TestParsePolicyPrincipalID(t *testing.T) {
	tests := []struct {
		name              string
		id                string
		wantPolicyID      string
		wantPrincipalID   string
		wantPrincipalType string
		wantErr           bool
		errContains       string
	}{
		{
			name:              "Valid policy:principal:type ID",
			id:                "policy-123:principal-456:user",
			wantPolicyID:      "policy-123",
			wantPrincipalID:   "principal-456",
			wantPrincipalType: "user",
			wantErr:           false,
		},
		{
			name:              "Valid with group type",
			id:                "policy-abc:principal-def:group",
			wantPolicyID:      "policy-abc",
			wantPrincipalID:   "principal-def",
			wantPrincipalType: "group",
			wantErr:           false,
		},
		{
			name:              "Valid with role type",
			id:                "123:456:role",
			wantPolicyID:      "123",
			wantPrincipalID:   "456",
			wantPrincipalType: "role",
			wantErr:           false,
		},
		{
			name:        "Missing parts",
			id:          "policy-123:principal-456",
			wantErr:     true,
			errContains: "expected 3 parts",
		},
		{
			name:              "ID with extra parts (handled by SplitN)",
			id:                "policy:principal:type:extra",
			wantPolicyID:      "policy",
			wantPrincipalID:   "principal",
			wantPrincipalType: "type:extra",
			wantErr:           false,
		},
		{
			name:        "Empty policy ID",
			id:          ":principal-456:user",
			wantErr:     true,
			errContains: "part 1 is empty",
		},
		{
			name:        "Empty principal ID",
			id:          "policy-123::user",
			wantErr:     true,
			errContains: "part 2 is empty",
		},
		{
			name:        "Empty principal type",
			id:          "policy-123:principal-456:",
			wantErr:     true,
			errContains: "part 3 is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyID, principalID, principalType, err := ParsePolicyPrincipalID(tt.id)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePolicyPrincipalID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}

			// Check error message
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParsePolicyPrincipalID(%q) error = %q, want error containing %q", tt.id, err.Error(), tt.errContains)
				}
			}

			// Check returned values
			if !tt.wantErr {
				if policyID != tt.wantPolicyID {
					t.Errorf("ParsePolicyPrincipalID(%q) policyID = %q, want %q", tt.id, policyID, tt.wantPolicyID)
				}
				if principalID != tt.wantPrincipalID {
					t.Errorf("ParsePolicyPrincipalID(%q) principalID = %q, want %q", tt.id, principalID, tt.wantPrincipalID)
				}
				if principalType != tt.wantPrincipalType {
					t.Errorf("ParsePolicyPrincipalID(%q) principalType = %q, want %q", tt.id, principalType, tt.wantPrincipalType)
				}
			}
		})
	}
}

// TestRoundTrip tests building and parsing composite IDs in a round-trip manner
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
	}{
		{
			name:  "Two-part ID",
			parts: []string{"policy-123", "database-456"},
		},
		{
			name:  "Three-part ID",
			parts: []string{"policy-abc", "principal-def", "user"},
		},
		{
			name:  "IDs with special characters",
			parts: []string{"policy_123-abc", "db-456_def"},
		},
		{
			name:  "UUID and numeric",
			parts: []string{"80b9f727-116d-4e6a-b682-f52fa8c25766", "193512"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build composite ID
			id := BuildCompositeID(tt.parts...)

			// Parse it back
			parts, err := ParseCompositeID(id, len(tt.parts))
			if err != nil {
				t.Fatalf("Round-trip failed: ParseCompositeID(%q, %d) returned error: %v", id, len(tt.parts), err)
			}

			// Verify parts match
			if len(parts) != len(tt.parts) {
				t.Fatalf("Round-trip failed: got %d parts, want %d", len(parts), len(tt.parts))
			}

			for i := range parts {
				if parts[i] != tt.parts[i] {
					t.Errorf("Round-trip failed: part[%d] = %q, want %q", i, parts[i], tt.parts[i])
				}
			}
		})
	}
}
