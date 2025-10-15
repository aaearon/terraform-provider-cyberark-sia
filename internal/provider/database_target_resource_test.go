// Package provider implements acceptance tests for database_target resource
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDatabaseTarget_basic tests basic CRUD lifecycle for database target resource
func TestAccDatabaseTarget_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseTargetConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "name", "test-postgres-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "database_version", "14.0.0"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "address", "postgres.example.com"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "port", "5432"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "authentication_method", "local"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.test", "cloud_provider", "on_premise"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cyberark_sia_database_target.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatabaseTarget_awsRDS tests AWS RDS PostgreSQL configuration
func TestAccDatabaseTarget_awsRDS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseTargetConfig_awsRDS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_target.aws_rds", "name", "aws-rds-postgres"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.aws_rds", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.aws_rds", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.aws_rds", "aws_region", "us-east-1"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.aws_rds", "aws_account_id", "123456789012"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.aws_rds", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseTarget_azureSQL tests Azure SQL Server configuration
func TestAccDatabaseTarget_azureSQL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseTargetConfig_azureSQL,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_target.azure_sql", "name", "azure-sqlserver"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.azure_sql", "database_type", "sqlserver"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.azure_sql", "cloud_provider", "azure"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.azure_sql", "authentication_method", "domain"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.azure_sql", "azure_tenant_id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.azure_sql", "azure_subscription_id"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.azure_sql", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseTarget_onPremise tests on-premise Oracle configuration
func TestAccDatabaseTarget_onPremise(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseTargetConfig_onPremise,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "name", "onprem-oracle-db"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "database_type", "oracle"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "database_version", "19.3.0"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "address", "oracle.internal.example.com"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "port", "1521"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "cloud_provider", "on_premise"),
					resource.TestCheckResourceAttrSet("cyberark_sia_database_target.oracle", "id"),
				),
			},
		},
	})
}

// TestAccDatabaseTarget_import tests ImportState functionality
func TestAccDatabaseTarget_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: testAccDatabaseTargetConfig_basic,
			},
			// Test import
			{
				ResourceName:      "cyberark_sia_database_target.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDatabaseTarget_multipleDatabaseTypes tests various database types
func TestAccDatabaseTarget_multipleDatabaseTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseTargetConfig_multipleTypes,
				Check: resource.ComposeAggregateTestCheckFunc(
					// PostgreSQL
					resource.TestCheckResourceAttr("cyberark_sia_database_target.postgres", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.postgres", "port", "5432"),
					// MySQL
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mysql", "database_type", "mysql"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mysql", "port", "3306"),
					// MariaDB
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mariadb", "database_type", "mariadb"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mariadb", "port", "3306"),
					// MongoDB
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mongodb", "database_type", "mongodb"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.mongodb", "port", "27017"),
					// Oracle
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "database_type", "oracle"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.oracle", "port", "1521"),
					// SQL Server
					resource.TestCheckResourceAttr("cyberark_sia_database_target.sqlserver", "database_type", "sqlserver"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.sqlserver", "port", "1433"),
					// Db2
					resource.TestCheckResourceAttr("cyberark_sia_database_target.db2", "database_type", "db2"),
					resource.TestCheckResourceAttr("cyberark_sia_database_target.db2", "port", "50000"),
				),
			},
		},
	})
}

// Test configurations

const testAccDatabaseTargetConfig_basic = `
resource "cyberark_sia_database_target" "test" {
  name              = "test-postgres-db"
  database_type     = "postgresql"
  database_version  = "14.0.0"
  address           = "postgres.example.com"
  port              = 5432
  database_name     = "testdb"
  authentication_method = "local"
  cloud_provider    = "on_premise"

  description = "Test PostgreSQL database"

  tags = {
    Environment = "test"
    ManagedBy   = "Terraform"
  }
}
`

const testAccDatabaseTargetConfig_awsRDS = `
resource "cyberark_sia_database_target" "aws_rds" {
  name              = "aws-rds-postgres"
  database_type     = "postgresql"
  database_version  = "14.7.0"
  address           = "mydb.abc123.us-east-1.rds.amazonaws.com"
  port              = 5432
  database_name     = "proddb"
  authentication_method = "local"

  cloud_provider = "aws"
  aws_region     = "us-east-1"
  aws_account_id = "123456789012"

  description = "AWS RDS PostgreSQL production database"

  tags = {
    Environment = "production"
    Cloud       = "AWS"
  }
}
`

const testAccDatabaseTargetConfig_azureSQL = `
resource "cyberark_sia_database_target" "azure_sql" {
  name              = "azure-sqlserver"
  database_type     = "sqlserver"
  database_version  = "13.0.0"
  address           = "sqlserver-prod.database.windows.net"
  port              = 1433
  database_name     = "appdb"
  authentication_method = "domain"

  cloud_provider        = "azure"
  azure_tenant_id       = "12345678-1234-1234-1234-123456789012"
  azure_subscription_id = "87654321-4321-4321-4321-210987654321"

  description = "Azure SQL Server production database"

  tags = {
    Environment = "production"
    Cloud       = "Azure"
  }
}
`

const testAccDatabaseTargetConfig_onPremise = `
resource "cyberark_sia_database_target" "oracle" {
  name              = "onprem-oracle-db"
  database_type     = "oracle"
  database_version  = "19.3.0"
  address           = "oracle.internal.example.com"
  port              = 1521
  database_name     = "PRODDB"
  authentication_method = "local"
  cloud_provider    = "on_premise"

  description = "On-premise Oracle production database"

  tags = {
    Environment = "production"
    Location    = "DataCenter-East"
  }
}
`

const testAccDatabaseTargetConfig_multipleTypes = `
resource "cyberark_sia_database_target" "postgres" {
  name              = "test-postgres"
  database_type     = "postgresql"
  database_version  = "14.0.0"
  address           = "postgres.example.com"
  port              = 5432
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "mysql" {
  name              = "test-mysql"
  database_type     = "mysql"
  database_version  = "8.0.0"
  address           = "mysql.example.com"
  port              = 3306
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "mariadb" {
  name              = "test-mariadb"
  database_type     = "mariadb"
  database_version  = "10.6.0"
  address           = "mariadb.example.com"
  port              = 3306
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "mongodb" {
  name              = "test-mongodb"
  database_type     = "mongodb"
  database_version  = "5.0.0"
  address           = "mongodb.example.com"
  port              = 27017
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "oracle" {
  name              = "test-oracle"
  database_type     = "oracle"
  database_version  = "19.3.0"
  address           = "oracle.example.com"
  port              = 1521
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "sqlserver" {
  name              = "test-sqlserver"
  database_type     = "sqlserver"
  database_version  = "13.0.0"
  address           = "sqlserver.example.com"
  port              = 1433
  authentication_method = "local"
  cloud_provider    = "on_premise"
}

resource "cyberark_sia_database_target" "db2" {
  name              = "test-db2"
  database_type     = "db2"
  database_version  = "11.5.0"
  address           = "db2.example.com"
  port              = 50000
  authentication_method = "local"
  cloud_provider    = "on_premise"
}
`
