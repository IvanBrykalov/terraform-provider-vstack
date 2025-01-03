// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVStackVMGetDataSource(t *testing.T) {
	// Define the Terraform configuration template for the data source.
	dataSourceConfigTemplate := `
variable "vm_id" {
  type        = number
}
data "vstack_vm_get" "test" {
  id = var.vm_id
}`
	// Combine the provider configuration template with the data source configuration template.
	dataSourceConfig := providerConfigTemplate + dataSourceConfigTemplate

	// Execute the test.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			{
				// Terraform configuration for the test step.
				Config: dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					// Ensure that all expected top-level attributes are set in the state.
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "id"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "name"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "description"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "cpus"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "ram"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "cpu_priority"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "boot_media"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "vcpu_class"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "os_type"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "os_profile"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "vdc_id"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "node"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "uefi"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "create_completed"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "locked"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "root_dataset"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "root_dataset_name"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "pool_selector"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "status"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "admin_status"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "oper_status"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "action"),

					// Ensure that guest nested attributes are set in the state.
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "guest.ram_used"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "guest.ram_balloon_performed"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "guest.ram_balloon_requested"),

					// Ensure that disks nested attributes are set in the state.
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.guid"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.slot"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.iops_limit"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.mbps_limit"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.sector_size.logical"),
					resource.TestCheckResourceAttrSet("data.vstack_vm_get.test", "disks.0.sector_size.physical"),
				),
			},
		},
	})
}
