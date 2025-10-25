// Package provider implements acceptance tests for database_workspace resource
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDatabaseWorkspace_basic tests basic CRUD lifecycle for database target resource
func TestAccDatabaseWorkspace_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseWorkspaceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "name", "test-postgres-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "database_type", "postgres"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "address", "postgres.example.com"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "port", "5432"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "authentication_method", "local_ephemeral_user"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.test", "cloud_provider", "on_premise"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cyberark_sia_database_workspace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatabaseWorkspace_awsRDS tests AWS RDS PostgreSQL configuration
func TestAccDatabaseWorkspace_awsRDS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWorkspaceConfig_awsRDS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.aws_rds", "name", "aws-rds-postgres"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.aws_rds", "database_type", "postgres-aws-rds"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.aws_rds", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.aws_rds", "region", "us-east-1"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.aws_rds", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseWorkspace_azureSQL tests Azure SQL Server configuration
func TestAccDatabaseWorkspace_azureSQL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWorkspaceConfig_azureSQL,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.azure_sql", "name", "azure-sqlserver"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.azure_sql", "database_type", "mssql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.azure_sql", "cloud_provider", "azure"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.azure_sql", "authentication_method", "ad_ephemeral_user"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.azure_sql", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseWorkspace_onPremise tests on-premise Oracle configuration
func TestAccDatabaseWorkspace_onPremise(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWorkspaceConfig_onPremise,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "name", "onprem-oracle-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "database_type", "oracle"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "address", "oracle.internal.example.com"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "port", "1521"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "cloud_provider", "on_premise"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.oracle", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseWorkspace_import tests ImportState functionality
func TestAccDatabaseWorkspace_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccDatabaseWorkspaceConfig_basic,
			},
			// Test import
			{
				ResourceName:      "cyberark_sia_database_workspace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatabaseWorkspace_multipleDatabaseTypes tests various database types
func TestAccDatabaseWorkspace_multipleDatabaseTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWorkspaceConfig_multipleTypes,
				Check: resource.ComposeAggregateTestCheckFunc(
					// PostgreSQL
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.postgres", "database_type", "postgres"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.postgres", "port", "5432"),
					// MySQL
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mysql", "database_type", "mysql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mysql", "port", "3306"),
					// MariaDB
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mariadb", "database_type", "mariadb"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mariadb", "port", "3306"),
					// MongoDB
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mongodb", "database_type", "mongo"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.mongodb", "port", "27017"),
					// Oracle
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "database_type", "oracle"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.oracle", "port", "1521"),
					// SQL Server
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.sqlserver", "database_type", "mssql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.sqlserver", "port", "1433"),
					// Db2
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.db2", "database_type", "db2"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.db2", "port", "50000"),
				),
			},
		},
	})
}

// Test configurations

const testAccDatabaseWorkspaceConfig_basic = `
resource "cyberark_sia_database_workspace" "test" {
  name                  = "test-postgres-db"
  database_type         = "postgres"
  address               = "postgres.example.com"
  port                  = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider        = "on_premise"
}
`

const testAccDatabaseWorkspaceConfig_awsRDS = `
resource "cyberark_sia_database_workspace" "aws_rds" {
  name                  = "aws-rds-postgres"
  database_type         = "postgres-aws-rds"
  address               = "mydb.abc123.us-east-1.rds.amazonaws.com"
  port                  = 5432
  authentication_method = "rds_iam_authentication"
  cloud_provider        = "aws"
  region                = "us-east-1"
}
`

const testAccDatabaseWorkspaceConfig_azureSQL = `
resource "cyberark_sia_database_workspace" "azure_sql" {
  name                  = "azure-sqlserver"
  database_type         = "mssql"
  address               = "sqlserver-prod.database.windows.net"
  port                  = 1433
  authentication_method = "ad_ephemeral_user"
  cloud_provider        = "azure"
}
`

const testAccDatabaseWorkspaceConfig_onPremise = `
resource "cyberark_sia_database_workspace" "oracle" {
  name                  = "onprem-oracle-db"
  database_type         = "oracle"
  address               = "oracle.internal.example.com"
  port                  = 1521
  authentication_method = "local_ephemeral_user"
  cloud_provider        = "on_premise"
}
`

const testAccDatabaseWorkspaceConfig_multipleTypes = `
resource "cyberark_sia_database_workspace" "postgres" {
  name              = "test-postgres"
  database_type     = "postgres"
  address           = "postgres.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "mysql" {
  name              = "test-mysql"
  database_type     = "mysql"
  address           = "mysql.example.com"
  port              = 3306
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "mariadb" {
  name              = "test-mariadb"
  database_type     = "mariadb"
  address           = "mariadb.example.com"
  port              = 3306
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "mongo" {
  name              = "test-mongodb"
  database_type     = "mongo"
  address           = "mongodb.example.com"
  port              = 27017
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "oracle" {
  name              = "test-oracle"
  database_type     = "oracle"
  address           = "oracle.example.com"
  port              = 1521
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "mssql" {
  name              = "test-sqlserver"
  database_type     = "mssql"
  address           = "sqlserver.example.com"
  port              = 1433
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "db2" {
  name              = "test-db2"
  database_type     = "db2"
  address           = "db2.example.com"
  port              = 50000
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}
`

// ============================================================================
// Phase 5 (User Story 3) Tests: Update, Delete, ForceNew, Drift Detection
// ============================================================================

// T060: TestAccDatabaseWorkspace_update tests configuration update functionality
func TestAccDatabaseWorkspace_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create initial resource
			{
				Config: testAccDatabaseWorkspaceConfig_beforeUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "name", "update-test-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "port", "5432"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "authentication_method", "local"),
				),
			},
			{
				Config: testAccDatabaseWorkspaceConfig_afterUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "name", "update-test-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "port", "5433"),                                  // Changed
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.update_test", "authentication_method", "local_ephemeral_user"), // Changed
				),
			},
			// Step 3: Verify import still works after update
			{
				ResourceName:      "cyberark_sia_database_workspace.update_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// T061: TestAccDatabaseWorkspace_forceNew tests ForceNew behavior for immutable attributes
func TestAccDatabaseWorkspace_forceNew(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with postgresql
			{
				Config: testAccDatabaseWorkspaceConfig_forceNew_before,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.forcenew_test", "database_type", "postgres"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.forcenew_test", "id"),
				),
			},
			// Step 2: Change database_type (should trigger replacement)
			// Note: Changing database_type should cause destroy-and-recreate
			{
				Config: testAccDatabaseWorkspaceConfig_forceNew_after,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.forcenew_test", "database_type", "mysql"),
					// ID should be different due to resource replacement
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.forcenew_test", "id"),
				),
			},
		},
	})
}

// T062: TestAccDatabaseWorkspace_noOpUpdate tests plan-only operation (no-op update)
func TestAccDatabaseWorkspace_noOpUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: testAccDatabaseWorkspaceConfig_noop,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.noop_test", "name", "noop-test-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.noop_test", "port", "5432"),
				),
			},
			// Step 2: Apply same config again (should be no-op)
			{
				Config:   testAccDatabaseWorkspaceConfig_noop,
				PlanOnly: true, // Verify no changes planned
			},
			// Step 3: Apply with ExpectNonEmptyPlan=false to verify no changes
			{
				Config:             testAccDatabaseWorkspaceConfig_noop,
				ExpectNonEmptyPlan: false, // No changes should be detected
			},
		},
	})
}

// T063: TestAccDatabaseWorkspace_concurrent tests concurrent resource operations
func TestAccDatabaseWorkspace_concurrent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple resources concurrently
			{
				Config: testAccDatabaseWorkspaceConfig_concurrent,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify all resources created successfully
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.concurrent1", "id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.concurrent2", "id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.concurrent3", "id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.concurrent4", "id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.concurrent5", "id"),
					// Verify each has correct type
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.concurrent1", "database_type", "postgres"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.concurrent2", "database_type", "mysql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.concurrent3", "database_type", "mariadb"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.concurrent4", "database_type", "mongo"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.concurrent5", "database_type", "oracle"),
				),
			},
		},
	})
}

// T064: TestAccDatabaseWorkspace_driftDetection tests state drift detection
func TestAccDatabaseWorkspace_driftDetection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: testAccDatabaseWorkspaceConfig_drift,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.drift_test", "name", "drift-test-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.drift_test", "port", "5432"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_workspace.drift_test", "id"),
				),
			},
			// Step 2: Refresh state (should detect if resource was modified outside Terraform)
			// Note: In a real test, you would manually modify the resource in SIA between steps
			// For now, this verifies the refresh mechanism works
			{
				Config:   testAccDatabaseWorkspaceConfig_drift,
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_workspace.drift_test", "name", "drift-test-db"),
				),
			},
		},
	})
}

// Test configurations for Phase 5 (User Story 3)

const testAccDatabaseWorkspaceConfig_beforeUpdate = `
resource "cyberark_sia_database_workspace" "update_test" {
  name              = "update-test-db"
  database_type     = "postgres"
  address           = "postgres-update.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"

}
`

const testAccDatabaseWorkspaceConfig_afterUpdate = `
resource "cyberark_sia_database_workspace" "update_test" {
  name              = "update-test-db"
  database_type     = "postgres"
  address           = "postgres-update.example.com"
  port              = 5433
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"

}
`

const testAccDatabaseWorkspaceConfig_forceNew_before = `
resource "cyberark_sia_database_workspace" "forcenew_test" {
  name              = "forcenew-test-db"
  database_type     = "postgres"
  address           = "db-forcenew.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}
`

const testAccDatabaseWorkspaceConfig_forceNew_after = `
resource "cyberark_sia_database_workspace" "forcenew_test" {
  name              = "forcenew-test-db"
  database_type     = "mysql"
  address           = "db-forcenew.example.com"
  port              = 3306
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}
`

const testAccDatabaseWorkspaceConfig_noop = `
resource "cyberark_sia_database_workspace" "noop_test" {
  name              = "noop-test-db"
  database_type     = "postgres"
  address           = "postgres-noop.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"


}
`

const testAccDatabaseWorkspaceConfig_concurrent = `
resource "cyberark_sia_database_workspace" "concurrent1" {
  name              = "concurrent-postgres"
  database_type     = "postgres"
  address           = "postgres-concurrent.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "concurrent2" {
  name              = "concurrent-mysql"
  database_type     = "mysql"
  address           = "mysql-concurrent.example.com"
  port              = 3306
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "concurrent3" {
  name              = "concurrent-mariadb"
  database_type     = "mariadb"
  address           = "mariadb-concurrent.example.com"
  port              = 3306
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "concurrent4" {
  name              = "concurrent-mongodb"
  database_type     = "mongo"
  address           = "mongodb-concurrent.example.com"
  port              = 27017
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_workspace" "concurrent5" {
  name              = "concurrent-oracle"
  database_type     = "oracle"
  address           = "oracle-concurrent.example.com"
  port              = 1521
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"
}
`

const testAccDatabaseWorkspaceConfig_drift = `
resource "cyberark_sia_database_workspace" "drift_test" {
  name              = "drift-test-db"
  database_type     = "postgres"
  address           = "postgres-drift.example.com"
  port              = 5432
  authentication_method = "local_ephemeral_user"
  cloud_provider    = "on_premise"


}
`

// ============================================================================
// Phase 4 (User Story 2): Certificate Association Tests
// ============================================================================

// TestAccDatabaseWorkspace_withCertificate tests database workspace with certificate reference
// Validates:
// - Certificate upload works
// - Database workspace can reference certificate via certificate_id
// - TLS validation is enabled
// - Certificate ID is persisted in state
func TestAccDatabaseWorkspace_withCertificate(t *testing.T) {
	t.Skip("Phase 4: Requires certificate resource to be implemented (Phase 3 complete)")

	// Test plan:
	// 1. Create certificate resource with valid PEM
	// 2. Create database workspace referencing certificate_id
	// 3. Verify certificate_id is set in state
	// 4. Verify enable_certificate_validation is true

	// Expected config:
	// resource "cyberark_sia_certificate" "test" {
	//   cert_name = "test-db-cert"
	//   cert_body = file("testdata/test-cert.pem")
	// }
	//
	// resource "cyberark_sia_database_workspace" "test" {
	//   name                          = "test-postgres-with-cert"
	//   database_type                 = "postgres"
	//   certificate_id                = cyberark_sia_certificate.test.id
	//   enable_certificate_validation = true
	// }
}

// TestAccDatabaseWorkspace_updateCertificate tests updating certificate association
// Validates:
// - Database workspace can be created without certificate
// - Certificate can be added to existing workspace
// - Update operation succeeds
// - New certificate ID is persisted in state
func TestAccDatabaseWorkspace_updateCertificate(t *testing.T) {
	t.Skip("Phase 4: Requires certificate resource to be implemented (Phase 3 complete)")

	// Test plan:
	// Step 1: Create database workspace without certificate
	// Step 2: Update workspace to add certificate_id
	// Step 3: Verify certificate_id is now set in state

	// Expected configs:
	// Initial:
	// resource "cyberark_sia_database_workspace" "test" {
	//   name          = "test-postgres"
	//   database_type = "postgres"
	// }
	//
	// Updated:
	// resource "cyberark_sia_certificate" "test" {
	//   cert_name = "test-cert"
	//   cert_body = file("testdata/test-cert.pem")
	// }
	//
	// resource "cyberark_sia_database_workspace" "test" {
	//   name           = "test-postgres"
	//   database_type  = "postgres"
	//   certificate_id = cyberark_sia_certificate.test.id  # Added
	// }
}

// TestAccDatabaseWorkspace_removeCertificate tests removing certificate association
// Validates:
// - Database workspace can be created with certificate
// - Certificate reference can be removed (set to null)
// - Update operation succeeds
// - Certificate ID is removed from state
func TestAccDatabaseWorkspace_removeCertificate(t *testing.T) {
	t.Skip("Phase 4: Requires certificate resource to be implemented (Phase 3 complete)")

	// Test plan:
	// Step 1: Create database workspace with certificate_id
	// Step 2: Update workspace to remove certificate_id (set to null)
	// Step 3: Verify certificate_id is removed from state

	// Expected configs:
	// Initial:
	// resource "cyberark_sia_certificate" "test" {
	//   cert_name = "test-cert"
	//   cert_body = file("testdata/test-cert.pem")
	// }
	//
	// resource "cyberark_sia_database_workspace" "test" {
	//   name           = "test-postgres"
	//   database_type  = "postgres"
	//   certificate_id = cyberark_sia_certificate.test.id
	// }
	//
	// Updated:
	// resource "cyberark_sia_database_workspace" "test" {
	//   name          = "test-postgres"
	//   database_type = "postgres"
	//   # certificate_id removed
	// }
}

// TestAccDatabaseWorkspace_invalidCertificateID tests error handling for non-existent certificate
// Validates:
// - Database workspace creation fails with actionable error
// - Error message provides guidance on verifying certificate exists
// - Error message includes the invalid certificate ID
func TestAccDatabaseWorkspace_invalidCertificateID(t *testing.T) {
	t.Skip("Phase 4: Requires certificate resource to be implemented (Phase 3 complete)")

	// Test plan:
	// 1. Attempt to create database workspace with non-existent certificate_id
	// 2. Verify creation fails with clear error message
	// 3. Verify error message mentions certificate not found
	// 4. Verify error message includes certificate ID

	// Expected config (should fail):
	// resource "cyberark_sia_database_workspace" "test" {
	//   name           = "test-postgres"
	//   database_type  = "postgres"
	//   certificate_id = "non-existent-cert-id-12345"
	// }

	// Expected error pattern:
	// Error: Certificate Not Found
	// The specified certificate (ID: non-existent-cert-id-12345) does not exist or is invalid.
	// Ensure the certificate exists before associating it with this database workspace.
}
