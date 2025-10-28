// Package validators provides custom validators for Terraform resources
package validators

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
)

var _ validator.String = databaseEngineValidator{}

// databaseEngineValidator validates that a string matches one of the SDK's valid database engine types.
type databaseEngineValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (v databaseEngineValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be one of the %d valid database engine types supported by the ARK SDK", len(dbmodels.DatabaseEngineTypes))
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v databaseEngineValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be one of the %d valid database engine types supported by the ARK SDK. "+
		"Valid types include generic variants (postgres, mysql, mariadb, mongo, oracle, mssql, db2) "+
		"and platform-specific variants (postgres-aws-rds, mysql-azure-managed, mongo-atlas-managed, etc.)",
		len(dbmodels.DatabaseEngineTypes))
}

// ValidateString performs the validation.
func (v databaseEngineValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there's nothing to validate
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	tflog.Trace(ctx, "Validating database engine type", map[string]interface{}{
		"value":                    value,
		"valid_engine_types_count": len(dbmodels.DatabaseEngineTypes),
	})

	// Check if value exists in SDK's list
	if !slices.Contains(dbmodels.DatabaseEngineTypes, value) {
		tflog.Warn(ctx, "Database engine validation failed", map[string]interface{}{
			"value":                    value,
			"valid_engine_types_count": len(dbmodels.DatabaseEngineTypes),
		})
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Database Engine Type",
			fmt.Sprintf("Value %q is not a valid database engine type. "+
				"Must be one of %d supported types: %s",
				value,
				len(dbmodels.DatabaseEngineTypes),
				formatEngineTypeExamples(),
			),
		)
	}
}

// formatEngineTypeExamples returns a formatted string of example engine types for error messages.
func formatEngineTypeExamples() string {
	// Show a few examples grouped by category
	examples := []string{
		"Generic: postgres, mysql, mariadb, mongo, oracle, mssql, sqlserver, db2",
		"AWS RDS: postgres-aws-rds, mysql-aws-rds, mariadb-aws-rds, oracle-aws-rds",
		"Azure: postgres-azure-managed, mysql-azure-managed, mssql-azure-managed",
		"Self-hosted: postgres-sh, mysql-sh, mariadb-sh, mongo-sh",
		"...and 45+ more variants",
	}
	return strings.Join(examples, "; ")
}

// DatabaseEngine returns a validator that checks if a string is a valid database engine type
// according to the ARK SDK's DatabaseEngineTypes list.
func DatabaseEngine() validator.String {
	return databaseEngineValidator{}
}
