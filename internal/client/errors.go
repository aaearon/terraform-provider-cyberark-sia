// Package client provides CyberArk SIA API client wrappers
package client

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// MapError converts ARK SDK errors to Terraform diagnostics with actionable guidance
// Returns nil if err is nil (caller should check before appending)
func MapError(err error, operation string) diag.Diagnostic {
	if err == nil {
		// Return an empty error diagnostic (won't be appended by caller check)
		return diag.NewErrorDiagnostic("", "")
	}

	errorMsg := err.Error()

	// Authentication errors
	if strings.Contains(errorMsg, "authentication failed") || strings.Contains(errorMsg, "invalid credentials") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Authentication Failed - %s", operation),
			fmt.Sprintf("Invalid client_id or client_secret.\n\n"+
				"Error: %s\n\n"+
				"Recommended actions:\n"+
				"1. Verify credentials in provider configuration\n"+
				"2. Check client ID format: client-id@cyberark.cloud.tenant-id\n"+
				"3. Ensure service account has SIA role memberships", errorMsg),
		)
	}

	// Permission errors
	if strings.Contains(errorMsg, "insufficient permissions") || strings.Contains(errorMsg, "forbidden") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Insufficient Permissions - %s", operation),
			fmt.Sprintf("ISPSS service account lacks required permissions.\n\n"+
				"Error: %s\n\n"+
				"Recommended action:\n"+
				"Verify service account has SIA Database Administrator role or equivalent", errorMsg),
		)
	}

	// Resource not found
	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "does not exist") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Resource Not Found - %s", operation),
			fmt.Sprintf("The requested resource was not found in SIA.\n\n"+
				"Error: %s\n\n"+
				"This may occur if:\n"+
				"- Resource was deleted outside Terraform\n"+
				"- Resource ID is incorrect\n\n"+
				"Run 'terraform refresh' to sync state", errorMsg),
		)
	}

	// Conflict/duplicate errors
	if strings.Contains(errorMsg, "already exists") || strings.Contains(errorMsg, "duplicate") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Resource Conflict - %s", operation),
			fmt.Sprintf("A resource with this identifier already exists.\n\n"+
				"Error: %s\n\n"+
				"Use 'terraform import' to manage the existing resource", errorMsg),
		)
	}

	// Validation errors
	if strings.Contains(errorMsg, "validation") || strings.Contains(errorMsg, "invalid") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Validation Failed - %s", operation),
			fmt.Sprintf("SIA API validation failed.\n\n"+
				"Error: %s\n\n"+
				"Check configuration values match SIA requirements", errorMsg),
		)
	}

	// Network/connectivity errors
	if strings.Contains(errorMsg, "connection refused") || strings.Contains(errorMsg, "timeout") {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Network Error - %s", operation),
			fmt.Sprintf("Unable to connect to SIA API.\n\n"+
				"Error: %s\n\n"+
				"Recommended actions:\n"+
				"1. Check network connectivity\n"+
				"2. Verify identity_url is correct\n"+
				"3. Check firewall rules", errorMsg),
		)
	}

	// Generic error with SDK message
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("SIA API Error - %s", operation),
		fmt.Sprintf("An error occurred communicating with SIA API.\n\n"+
			"Error: %s", errorMsg),
	)
}
