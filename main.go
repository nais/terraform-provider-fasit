package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/nais/terraform-provider-fasit/internal/provider"
)

// Run "mise run generate" to format example terraform files and generate the docs for the registry/website.
//
// Keep docs generation on Terraform (not OpenTofu) to avoid registry host mismatches during schema export.

//go:generate mise x terraform@1.10.5 -- terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --tf-version 1.10.5

var version string = "dev" // goreleaser can also pass the specific commit if you want

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "tfregistry.cloud.nais.io/nais/fasit",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
