package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	dbmodels "github.com/cyberark/ark-sdk-golang/pkg/services/sia/workspaces/db/models"
)

func TestDatabaseEngineValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		// Generic engines
		{
			name:      "valid postgres",
			value:     types.StringValue("postgres"),
			expectErr: false,
		},
		{
			name:      "valid mysql",
			value:     types.StringValue("mysql"),
			expectErr: false,
		},
		{
			name:      "valid mariadb",
			value:     types.StringValue("mariadb"),
			expectErr: false,
		},
		{
			name:      "valid mongo",
			value:     types.StringValue("mongo"),
			expectErr: false,
		},
		{
			name:      "valid oracle",
			value:     types.StringValue("oracle"),
			expectErr: false,
		},
		{
			name:      "valid mssql",
			value:     types.StringValue("mssql"),
			expectErr: false,
		},
		{
			name:      "valid sqlserver",
			value:     types.StringValue("sqlserver"),
			expectErr: false,
		},
		{
			name:      "valid db2",
			value:     types.StringValue("db2"),
			expectErr: false,
		},

		// AWS RDS variants
		{
			name:      "valid postgres-aws-rds",
			value:     types.StringValue("postgres-aws-rds"),
			expectErr: false,
		},
		{
			name:      "valid mysql-aws-rds",
			value:     types.StringValue("mysql-aws-rds"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-aws-rds",
			value:     types.StringValue("mariadb-aws-rds"),
			expectErr: false,
		},
		{
			name:      "valid oracle-aws-rds",
			value:     types.StringValue("oracle-aws-rds"),
			expectErr: false,
		},
		{
			name:      "valid mssql-aws-rds",
			value:     types.StringValue("mssql-aws-rds"),
			expectErr: false,
		},
		{
			name:      "valid db2-aws-rds",
			value:     types.StringValue("db2-aws-rds"),
			expectErr: false,
		},

		// AWS Aurora variants
		{
			name:      "valid postgres-aws-aurora",
			value:     types.StringValue("postgres-aws-aurora"),
			expectErr: false,
		},
		{
			name:      "valid mysql-aws-aurora",
			value:     types.StringValue("mysql-aws-aurora"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-aws-aurora",
			value:     types.StringValue("mariadb-aws-aurora"),
			expectErr: false,
		},

		// AWS VM variants
		{
			name:      "valid postgres-aws-vm",
			value:     types.StringValue("postgres-aws-vm"),
			expectErr: false,
		},
		{
			name:      "valid mysql-aws-vm",
			value:     types.StringValue("mysql-aws-vm"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-aws-vm",
			value:     types.StringValue("mariadb-aws-vm"),
			expectErr: false,
		},
		{
			name:      "valid oracle-aws-vm",
			value:     types.StringValue("oracle-aws-vm"),
			expectErr: false,
		},
		{
			name:      "valid mssql-aws-ec2",
			value:     types.StringValue("mssql-aws-ec2"),
			expectErr: false,
		},

		// Azure managed variants
		{
			name:      "valid postgres-azure-managed",
			value:     types.StringValue("postgres-azure-managed"),
			expectErr: false,
		},
		{
			name:      "valid mysql-azure-managed",
			value:     types.StringValue("mysql-azure-managed"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-azure-managed",
			value:     types.StringValue("mariadb-azure-managed"),
			expectErr: false,
		},
		{
			name:      "valid mssql-azure-managed",
			value:     types.StringValue("mssql-azure-managed"),
			expectErr: false,
		},

		// Azure VM variants
		{
			name:      "valid postgres-azure-vm",
			value:     types.StringValue("postgres-azure-vm"),
			expectErr: false,
		},
		{
			name:      "valid mysql-azure-vm",
			value:     types.StringValue("mysql-azure-vm"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-azure-vm",
			value:     types.StringValue("mariadb-azure-vm"),
			expectErr: false,
		},
		{
			name:      "valid mssql-azure-vm",
			value:     types.StringValue("mssql-azure-vm"),
			expectErr: false,
		},

		// Self-hosted variants
		{
			name:      "valid postgres-sh",
			value:     types.StringValue("postgres-sh"),
			expectErr: false,
		},
		{
			name:      "valid mysql-sh",
			value:     types.StringValue("mysql-sh"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-sh",
			value:     types.StringValue("mariadb-sh"),
			expectErr: false,
		},
		{
			name:      "valid mongo-sh",
			value:     types.StringValue("mongo-sh"),
			expectErr: false,
		},
		{
			name:      "valid oracle-sh",
			value:     types.StringValue("oracle-sh"),
			expectErr: false,
		},
		{
			name:      "valid mssql-sh",
			value:     types.StringValue("mssql-sh"),
			expectErr: false,
		},
		{
			name:      "valid sqlserver-sh",
			value:     types.StringValue("sqlserver-sh"),
			expectErr: false,
		},
		{
			name:      "valid db2-sh",
			value:     types.StringValue("db2-sh"),
			expectErr: false,
		},

		// Self-hosted VM variants
		{
			name:      "valid postgres-sh-vm",
			value:     types.StringValue("postgres-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid mysql-sh-vm",
			value:     types.StringValue("mysql-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid mariadb-sh-vm",
			value:     types.StringValue("mariadb-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid mongo-sh-vm",
			value:     types.StringValue("mongo-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid oracle-sh-vm",
			value:     types.StringValue("oracle-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid mssql-sh-vm",
			value:     types.StringValue("mssql-sh-vm"),
			expectErr: false,
		},
		{
			name:      "valid db2-sh-vm",
			value:     types.StringValue("db2-sh-vm"),
			expectErr: false,
		},

		// Atlas managed variants
		{
			name:      "valid mongo-atlas-managed",
			value:     types.StringValue("mongo-atlas-managed"),
			expectErr: false,
		},

		// AWS DocumentDB
		{
			name:      "valid mongo-aws-docdb",
			value:     types.StringValue("mongo-aws-docdb"),
			expectErr: false,
		},

		// Oracle-specific variants
		{
			name:      "valid oracle-ee",
			value:     types.StringValue("oracle-ee"),
			expectErr: false,
		},
		{
			name:      "valid oracle-ee-cdb",
			value:     types.StringValue("oracle-ee-cdb"),
			expectErr: false,
		},
		{
			name:      "valid oracle-se2",
			value:     types.StringValue("oracle-se2"),
			expectErr: false,
		},
		{
			name:      "valid oracle-se2-cdb",
			value:     types.StringValue("oracle-se2-cdb"),
			expectErr: false,
		},

		// SQL Server-specific variants
		{
			name:      "valid custom-sqlserver-ee",
			value:     types.StringValue("custom-sqlserver-ee"),
			expectErr: false,
		},
		{
			name:      "valid custom-sqlserver-se",
			value:     types.StringValue("custom-sqlserver-se"),
			expectErr: false,
		},
		{
			name:      "valid custom-sqlserver-web",
			value:     types.StringValue("custom-sqlserver-web"),
			expectErr: false,
		},

		// Invalid engines - common misspellings
		{
			name:      "invalid postgresql (should be postgres)",
			value:     types.StringValue("postgresql"),
			expectErr: true,
		},
		{
			name:      "invalid mongodb (should be mongo)",
			value:     types.StringValue("mongodb"),
			expectErr: true,
		},
		{
			name:      "invalid sql-server (should be sqlserver)",
			value:     types.StringValue("sql-server"),
			expectErr: true,
		},
		{
			name:      "invalid ms-sql (should be mssql)",
			value:     types.StringValue("ms-sql"),
			expectErr: true,
		},

		// Invalid engines - case variations
		{
			name:      "invalid POSTGRES (case sensitive)",
			value:     types.StringValue("POSTGRES"),
			expectErr: true,
		},
		{
			name:      "invalid MySQL (case sensitive)",
			value:     types.StringValue("MySQL"),
			expectErr: true,
		},

		// Invalid engines - generic
		{
			name:      "invalid engine",
			value:     types.StringValue("invalid"),
			expectErr: true,
		},
		{
			name:      "invalid unknown",
			value:     types.StringValue("unknown"),
			expectErr: true,
		},
		{
			name:      "empty string",
			value:     types.StringValue(""),
			expectErr: true,
		},

		// Null/unknown values (should pass - skip validation)
		{
			name:      "null value (allowed)",
			value:     types.StringNull(),
			expectErr: false,
		},
		{
			name:      "unknown value (allowed)",
			value:     types.StringUnknown(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := DatabaseEngine()
			req := validator.StringRequest{
				Path:        path.Root("database_engine"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectErr {
				t.Errorf("DatabaseEngine() hasError = %v, expectErr %v", hasError, tt.expectErr)
				if hasError {
					t.Logf("Diagnostics: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestDatabaseEngineValidator_Description(t *testing.T) {
	v := DatabaseEngine()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	markdownDesc := v.MarkdownDescription(ctx)
	if markdownDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}
}

// TestDatabaseEngineValidator_SDKCoverage verifies that the ARK SDK provides
// a reasonable number of database engine types (50+).
func TestDatabaseEngineValidator_SDKCoverage(t *testing.T) {
	engineCount := len(dbmodels.DatabaseEngineTypes)
	if engineCount < 50 {
		t.Errorf("Expected at least 50 database engine types, got %d", engineCount)
	}

	t.Logf("ARK SDK provides %d database engine types", engineCount)
}
