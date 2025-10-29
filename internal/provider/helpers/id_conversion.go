// Package helpers provides shared utility functions for provider resources
package helpers

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ConvertDatabaseIDToInt converts a database ID string to integer with proper error handling
// Used by database_workspace_resource, database_policy_resource, and policy_database_assignment_resource
func ConvertDatabaseIDToInt(id string, diagnostics *diag.Diagnostics, attrPath path.Path) (int, bool) {
	databaseIDInt, err := strconv.Atoi(id)
	if err != nil {
		diagnostics.AddAttributeError(
			attrPath,
			"Invalid Database ID",
			fmt.Sprintf("Database workspace ID must be a valid integer: %s", err.Error()),
		)
		return 0, false
	}
	return databaseIDInt, true
}

// ConvertIntToString converts an integer ID to string (common pattern for SDK responses)
func ConvertIntToString(id int) string {
	return strconv.Itoa(id)
}
