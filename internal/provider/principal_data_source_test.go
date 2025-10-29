// Package provider implements acceptance tests for principal_data_source
package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPrincipalDataSource_CloudUser tests Cloud Directory (CDS) user lookup
// US1: Cloud Directory User Lookup
func TestAccPrincipalDataSource_CloudUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_CloudUser,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.cloud_user", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.cloud_user", "principal_type", "USER"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.cloud_user", "directory_name", "CyberArk Cloud Directory"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.cloud_user", "directory_id"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.cloud_user", "display_name"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.cloud_user", "email"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_FederatedUser tests Federated Directory (FDS) user lookup
// US2: Federated Directory User Lookup
func TestAccPrincipalDataSource_FederatedUser(t *testing.T) {
	// Valid federated user: tim.schindler@cyberiam.com

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_FederatedUser,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.federated_user", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.federated_user", "principal_type", "USER"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.federated_user", "directory_name"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.federated_user", "directory_id"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_Group tests group lookup
// US3: Group Lookup
func TestAccPrincipalDataSource_Group(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_Group,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.group", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.group", "principal_type", "GROUP"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.group", "directory_name"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.group", "directory_id"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.group", "display_name"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_TypeFilter tests optional type parameter filtering
// US3: Type Filter Validation
func TestAccPrincipalDataSource_TypeFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_TypeFilter,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.filtered_user", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.filtered_user", "principal_type", "USER"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_ADUser tests Active Directory (AdProxy) user lookup
// US4: Active Directory User Lookup
func TestAccPrincipalDataSource_ADUser(t *testing.T) {
	// Valid AD user: SchindlerT@cyberiam.tech

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_ADUser,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.ad_user", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.ad_user", "principal_type", "USER"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.ad_user", "directory_name"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.ad_user", "directory_id"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_NotFound tests error handling when principal doesn't exist
// US5: Error Handling - Principal Not Found
func TestAccPrincipalDataSource_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPrincipalDataSourceConfig_NotFound,
				ExpectError: regexp.MustCompile("Principal Not Found"),
			},
		},
	})
}

// TestAccPrincipalDataSource_Role tests ROLE principal type lookup
func TestAccPrincipalDataSource_Role(t *testing.T) {
	// Valid role: System Administrator

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_Role,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.role", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.role", "principal_type", "ROLE"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.role", "directory_name"),
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.role", "directory_id"),
				),
			},
		},
	})
}

// TestAccPrincipalDataSource_WithPolicyAssignment tests integration with policy assignment
func TestAccPrincipalDataSource_WithPolicyAssignment(t *testing.T) {
	t.Skip("Integration test - requires full database + policy setup - run manually")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalDataSourceConfig_WithPolicyAssignment,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Principal lookup
					resource.TestCheckResourceAttrSet("data.cyberarksia_principal.policy_user", "id"),
					resource.TestCheckResourceAttr("data.cyberarksia_principal.policy_user", "principal_type", "USER"),
					// Policy assignment using principal data
					resource.TestCheckResourceAttrSet("cyberark_sia_database_policy_principal_assignment.test", "id"),
				),
			},
		},
	})
}

// Test configurations

const testAccPrincipalDataSourceConfig_CloudUser = `
provider "cyberarksia" {}

data "cyberarksia_principal" "cloud_user" {
  name = "tim.schindler@cyberark.cloud.40562"
}
`

const testAccPrincipalDataSourceConfig_FederatedUser = `
provider "cyberarksia" {}

data "cyberarksia_principal" "federated_user" {
  name = "tim.schindler@cyberiam.com"
}
`

const testAccPrincipalDataSourceConfig_Group = `
provider "cyberarksia" {}

data "cyberarksia_principal" "group" {
  name = "CyberArk Guardians"
}
`

const testAccPrincipalDataSourceConfig_TypeFilter = `
provider "cyberarksia" {}

data "cyberarksia_principal" "filtered_user" {
  name = "tim.schindler@cyberark.cloud.40562"
  type = "USER"
}
`

const testAccPrincipalDataSourceConfig_ADUser = `
provider "cyberarksia" {}

data "cyberarksia_principal" "ad_user" {
  name = "SchindlerT@cyberiam.tech"
}
`

const testAccPrincipalDataSourceConfig_NotFound = `
provider "cyberarksia" {}

data "cyberarksia_principal" "nonexistent" {
  name = "nonexistent.user.does.not.exist@invalid.domain.test"
}
`

const testAccPrincipalDataSourceConfig_Role = `
provider "cyberarksia" {}

data "cyberarksia_principal" "role" {
  name = "System Administrator"
  type = "ROLE"
}
`

const testAccPrincipalDataSourceConfig_WithPolicyAssignment = `
data "cyberarksia_principal" "policy_user" {
  name = "tim.schindler@cyberark.cloud.40562"
}

data "cyberark_sia_access_policy" "existing_policy" {
  name = "Production Database Policy"
}

resource "cyberark_sia_database_policy_principal_assignment" "test" {
  policy_id               = data.cyberark_sia_access_policy.existing_policy.id
  principal_id            = data.cyberarksia_principal.policy_user.id
  principal_type          = data.cyberarksia_principal.policy_user.principal_type
  principal_name          = data.cyberarksia_principal.policy_user.name
  source_directory_name   = data.cyberarksia_principal.policy_user.directory_name
  source_directory_id     = data.cyberarksia_principal.policy_user.directory_id
}
`
