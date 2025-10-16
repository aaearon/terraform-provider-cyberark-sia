// Package main is the entry point for the CyberArk SIA Terraform provider
package main

import (
	"context"
	"flag"
	"log"

	"github.com/aaearon/terraform-provider-cyberark-sia/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "local/aaearon/cyberark-sia",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
