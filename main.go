package main

import (
	"flag"

	"contabo.com/terraform-provider-contabo/client"
	"contabo.com/terraform-provider-contabo/contabo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Generate the Terraform provider documentation using `tfplugindocs`.
// --provider-name must match the resource prefix ("contabo_*"); otherwise
// newer tfplugindocs derives it from the repo dir and fails schema lookup.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name contabo

// Set by goreleaser.
var (
	version = "dev"
	commit  = "none"
)

// providerAddress is the registry address of this distribution.
const providerAddress = "registry.terraform.io/hayone1/contabo"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	client.UserAgent = "terraform-provider-contabo/" + version + " (" + providerAddress + "; commit " + commit + ")"

	plugin.Serve(&plugin.ServeOpts{
		Debug:        debug,
		ProviderAddr: providerAddress,
		ProviderFunc: func() *schema.Provider {
			return contabo.Provider()
		},
	})
}
