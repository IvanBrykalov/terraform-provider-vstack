package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// providerConfigTemplate is a template for the provider configuration using environment variables.
const providerConfigTemplate = `
variable "host" {
  type        = string
  nullable = false
  sensitive   = true
}
variable "username" {
  type        = string
  nullable = false
  sensitive   = true
}
variable "password" {
  type        = string
  nullable = false
  sensitive   = true
}
provider "vstack" {
  username = var.username
  password = var.password
  host     = var.host
}
`

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"vstack": providerserver.NewProtocol6WithError(New("test")()),
	}
)
