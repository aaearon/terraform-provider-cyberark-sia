// Package helpers provides shared utility functions for provider resources
package helpers

import (
	"fmt"
	"strings"
)

// BuildCompositeID creates a composite ID from two parts
// Used by policy_database_assignment (policy:database) and policy_principal_assignment (policy:principal:type)
func BuildCompositeID(parts ...string) string {
	return strings.Join(parts, ":")
}

// ParseCompositeID splits a composite ID into its parts
// Returns error if ID format is invalid
func ParseCompositeID(id string, expectedParts int) ([]string, error) {
	parts := strings.SplitN(id, ":", expectedParts)
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("invalid composite ID format: expected %d parts separated by ':', got %d parts in '%s'",
			expectedParts, len(parts), id)
	}

	// Validate no empty parts
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid composite ID format: part %d is empty in '%s'", i+1, id)
		}
	}

	return parts, nil
}

// ParsePolicyDatabaseID parses a policy:database composite ID
func ParsePolicyDatabaseID(id string) (policyID, databaseID string, err error) {
	parts, err := ParseCompositeID(id, 2)
	if err != nil {
		return "", "", err
	}
	return parts[0], parts[1], nil
}

// ParsePolicyPrincipalID parses a policy:principal:type composite ID
func ParsePolicyPrincipalID(id string) (policyID, principalID, principalType string, err error) {
	parts, err := ParseCompositeID(id, 3)
	if err != nil {
		return "", "", "", err
	}
	return parts[0], parts[1], parts[2], nil
}
