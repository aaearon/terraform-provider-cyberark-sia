// Package provider implements acceptance tests for secret resource
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ============================================================================
// Phase 4 (User Story 2) Tests: Strong Account CRUD Lifecycle
// ============================================================================

// T039: TestAccSecret_basic tests basic CRUD lifecycle for strong account resource
func TestAccSecret_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.test", "name", "test-strong-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.test", "authentication_type", "local"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.test", "username", "db_admin"),
					resource.TestCheckResourceAttrSet("cyberark_sia_secret.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cyberark_sia_secret.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Password is sensitive and won't be in state
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// T040: TestAccSecret_localAuth tests local authentication strong account
func TestAccSecret_localAuth(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_localAuth,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.local", "name", "local-auth-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.local", "authentication_type", "local"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.local", "username", "postgres_admin"),
					resource.TestCheckResourceAttrSet("cyberark_sia_secret.local", "id"),
				),
			},
		},
	})
}

// T041: TestAccSecret_domainAuth tests Active Directory authentication strong account
func TestAccSecret_domainAuth(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_domainAuth,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.domain", "name", "domain-auth-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.domain", "authentication_type", "domain"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.domain", "username", "CORP\\sqladmin"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.domain", "domain", "corp.example.com"),
					resource.TestCheckResourceAttrSet("cyberark_sia_secret.domain", "id"),
				),
			},
		},
	})
}

// T042: TestAccSecret_awsIAM tests AWS IAM authentication strong account
func TestAccSecret_awsIAM(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_awsIAM,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.aws_iam", "name", "aws-iam-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.aws_iam", "authentication_type", "aws_iam"),
					resource.TestCheckResourceAttrSet("cyberark_sia_secret.aws_iam", "id"),
				),
			},
		},
	})
}

// T043: TestAccSecret_credentialUpdate tests credential rotation/update
func TestAccSecret_credentialUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with initial credentials
			{
				Config: testAccSecretConfig_credentialsBefore,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.rotation_test", "name", "rotation-test-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.rotation_test", "username", "initial_user"),
				),
			},
			// Step 2: Update credentials (password and username)
			{
				Config: testAccSecretConfig_credentialsAfter,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.rotation_test", "name", "rotation-test-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.rotation_test", "username", "updated_user"),
				),
			},
		},
	})
}

// T044: TestAccSecret_import tests ImportState functionality
func TestAccSecret_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccSecretConfig_basic,
			},
			// Test import
			{
				ResourceName:      "cyberark_sia_secret.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Password is sensitive and won't be in state
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// ============================================================================
// Phase 5 (User Story 3) Tests: Strong Account Credential Update
// ============================================================================

// T065: TestAccSecret_updateCredentials tests strong account credential rotation
func TestAccSecret_updateCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with initial password
			{
				Config: testAccSecretConfig_updateBefore,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "name", "update-test-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "username", "db_user"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "description", "Initial credentials"),
				),
			},
			// Step 2: Rotate password (credential update)
			{
				Config: testAccSecretConfig_updateAfter,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "name", "update-test-account"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "username", "db_user"),
					resource.TestCheckResourceAttr("cyberark_sia_secret.update_test", "description", "Rotated credentials"),
				),
			},
			// Step 3: Verify import still works after update
			{
				ResourceName:      "cyberark_sia_secret.update_test",
				ImportState:       true,
				ImportStateVerify: true,
				// Password is sensitive and won't be in state
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// ============================================================================
// Test Configurations
// ============================================================================

const testAccSecretConfig_basic = `
resource "cyberark_sia_database_workspace" "test" {
  name          = "test-db-for-strong-account"
  database_type = "postgresql"
  address       = "postgres.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "test" {
  name               = "test-strong-account"
  authentication_type = "local"
  username           = "db_admin"
  password           = "InitialPassword123!"

  description = "Test strong account"

  tags = {
    Environment = "test"
    ManagedBy   = "Terraform"
  }
}
`

const testAccSecretConfig_localAuth = `
resource "cyberark_sia_database_workspace" "postgres" {
  name          = "postgres-db-local"
  database_type = "postgresql"
  address       = "postgres-local.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "local" {
  name               = "local-auth-account"
  authentication_type = "local"
  username           = "postgres_admin"
  password           = "SecurePassword456!"

  description = "Local authentication strong account for PostgreSQL"

  tags = {
    Environment = "test"
    AuthType    = "local"
  }
}
`

const testAccSecretConfig_domainAuth = `
resource "cyberark_sia_database_workspace" "sqlserver" {
  name          = "sqlserver-db-domain"
  database_type = "sqlserver"
  address       = "sqlserver-domain.example.com"
  port          = 1433
}

resource "cyberark_sia_secret" "domain" {
  name               = "domain-auth-account"
  authentication_type = "domain"
  username           = "CORP\\sqladmin"
  password           = "DomainPassword789!"
  domain             = "corp.example.com"

  description = "Active Directory authentication strong account for SQL Server"

  tags = {
    Environment = "test"
    AuthType    = "domain"
  }
}
`

const testAccSecretConfig_awsIAM = `
resource "cyberark_sia_database_workspace" "rds" {
  name          = "rds-db-iam"
  database_type = "postgresql"
  address       = "mydb.abc123.us-east-1.rds.amazonaws.com"
  port          = 5432
  cloud_provider = "aws"
  region        = "us-east-1"
}

resource "cyberark_sia_secret" "aws_iam" {
  name               = "aws-iam-account"
  authentication_type = "aws_iam"
  aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
  aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

  description = "AWS IAM authentication strong account for RDS"

  tags = {
    Environment = "test"
    AuthType    = "aws_iam"
  }
}
`

const testAccSecretConfig_credentialsBefore = `
resource "cyberark_sia_database_workspace" "rotation" {
  name          = "rotation-test-db"
  database_type = "postgresql"
  address       = "postgres-rotation.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "rotation_test" {
  name               = "rotation-test-account"
  authentication_type = "local"
  username           = "initial_user"
  password           = "InitialPassword123!"

  description = "Account for credential rotation testing"

  tags = {
    Environment = "test"
    Phase       = "before-rotation"
  }
}
`

const testAccSecretConfig_credentialsAfter = `
resource "cyberark_sia_database_workspace" "rotation" {
  name          = "rotation-test-db"
  database_type = "postgresql"
  address       = "postgres-rotation.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "rotation_test" {
  name               = "rotation-test-account"
  authentication_type = "local"
  username           = "updated_user"
  password           = "RotatedPassword456!"

  description = "Account for credential rotation testing"

  tags = {
    Environment = "test"
    Phase       = "after-rotation"
  }
}
`

const testAccSecretConfig_updateBefore = `
resource "cyberark_sia_database_workspace" "update" {
  name          = "update-test-db"
  database_type = "postgresql"
  address       = "postgres-update.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "update_test" {
  name               = "update-test-account"
  authentication_type = "local"
  username           = "db_user"
  password           = "InitialPassword123!"

  description = "Initial credentials"

  tags = {
    Environment = "test"
    Phase       = "before-update"
  }
}
`

const testAccSecretConfig_updateAfter = `
resource "cyberark_sia_database_workspace" "update" {
  name          = "update-test-db"
  database_type = "postgresql"
  address       = "postgres-update.example.com"
  port          = 5432
}

resource "cyberark_sia_secret" "update_test" {
  name               = "update-test-account"
  authentication_type = "local"
  username           = "db_user"
  password           = "RotatedPassword456!"

  description = "Rotated credentials"

  tags = {
    Environment = "test"
    Phase       = "after-update"
  }
}
`
