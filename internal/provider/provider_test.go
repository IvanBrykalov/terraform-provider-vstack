// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
)

// providerConfigTemplate is a template for the provider configuration using environment variables.

const providerConfigTemplate = `
variable "host" {
  type        = string
  sensitive   = true
}
variable "username" {
  type        = string
  sensitive   = true
}
variable "password" {
  type        = string
  sensitive   = true
}
provider "vstack" {
  username = var.username
  password = var.password
  host     = var.host
}

variable "vdc_id" {
  type        = number
}

variable network_id {
  type        = number
}

variable "os_profile" {
  type        = string
  default     = "4001"
}

variable ip_address {
  type        = string
  default     = "192.168.0.100"
}

variable "pool_selector" {
  type        = string
  default     = "14061357726568775332"
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

// TestMain is the entry point for all tests in this package. It validates that
// all required environment variables are present. If any are missing, the tests
// will fail immediately (exit code 1).
func TestMain(m *testing.M) {
	requiredVars := []string{
		"TF_VAR_host",
		"TF_VAR_username",
		"TF_VAR_password",
		"TF_VAR_vdc_id",
		"TF_VAR_network_id",
	}

	var missingVars []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missingVars = append(missingVars, v)
		}
	}

	if len(missingVars) > 0 {
		fmt.Printf("ERROR: Missing required environment variables: %v\n", missingVars)
		// Approach #1: Exit with a non-zero status, marking tests as failed.
		os.Exit(1)
	}

	// If all variables are set, proceed with the tests.
	code := m.Run()
	os.Exit(code)
}
