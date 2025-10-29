// Package provider implements the CyberArk SIA Terraform provider
package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cyberarksia": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that the required environment variables are set
// for acceptance tests. This function is called before each acceptance test.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Check that TF_ACC is set to enable acceptance tests
	if os.Getenv(EnvTFAcc) == "" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}

	// Validate required environment variables
	for _, envVar := range TestAccPreCheckVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}
