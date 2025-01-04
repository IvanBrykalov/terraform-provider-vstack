// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVStackVMProfileDataSource(t *testing.T) {
	// Terraform configuration snippet for the vstack_vm_profile data source.
	dataSourceConfigTemplate := `
data "vstack_vm_profile" "test" {}
`

	// Combine the standard provider config with our data source config.
	dataSourceConfig := providerConfigTemplate + dataSourceConfigTemplate

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Provide the complete Terraform config for this test step.
				Config: dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					// Verify that at least one OS type exists in the state.
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.id"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.name"),

					// Verify that at least one profile is defined under that OS type.
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.profiles.0.id"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.profiles.0.name"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.profiles.0.description"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_profile.test", "os_types.0.profiles.0.min_size"),
				),
			},
		},
	})
}
