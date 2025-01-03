// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

// TestAccVStackVM tests the VStack VM resource, including Create, Update, and Delete steps.
func TestAccVStackVM(t *testing.T) {
	// Define the Terraform configuration template for the resource.
	resourceConfigTemplate := `
variable "vdc_id" {
  type        = number
}

resource "vstack_vm" "test_vm2" {
  name          = "test-vm2"
  description   = "This is a test VM for testing purposes."
  cpus          = 1
  ram           = 2048
  cpu_priority  = 10
  boot_media    = 0
  vcpu_class    = 1
  os_type       = 6
  os_profile    = "4001"
  vdc_id        = var.vdc_id
  pool_selector = "14061357726568775332"

  action = "start"

  disks = [
    {
      size       = 20
      slot       = 1
      label      = "Primary Disk"
      iops_limit = 200
      mbps_limit = 256
    }
  ]

  # Guest params
  guest = {
    hostname           = "test_vm2"
    ssh_password_auth  = 1

    # User config
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."
        ]
        password = "rootpassword"
      }

      user = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."        ]
        password = "userpassword"
      }
    }

    # DNS
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "test.local"
    }

    # # Boot commands
    boot_cmds = ["swapoff -a"]
    
    # Commands after bootstrap
    run_cmds = ["touch /root/test"]
  }
}
`

	// Define the updated Terraform configuration template for the Update step.
	resourceConfigUpdateTemplate := `
variable "vdc_id" {
  type        = number
}

resource "vstack_vm" "test_vm2" {
  name          = "test-vm2"
  description   = "This is a test VM for testing purposes."
  cpus          = 2
  ram           = 4096
  cpu_priority  = 10
  boot_media    = 0
  vcpu_class    = 1
  os_type       = 6
  os_profile    = "4001"
  vdc_id        = var.vdc_id
  pool_selector = "14061357726568775332"

  action = "stop"

  disks = [
    {
      size       = 25
      slot       = 1
      label      = "Primary Disk-1"
      iops_limit = 0
      mbps_limit = 0
    },
	{
      size       = 20
      slot       = 2
      label      = "Secondary Disk-2"
      iops_limit = 0
      mbps_limit = 0
    }
  ]
  guest = {
    hostname           = "test_vm2"
    ssh_password_auth  = 1
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."
        ]
        password = "rootpassword"
      }

      user = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."        ]
        password = "userpassword"
      }
    }
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "test.local"
    }
    boot_cmds = ["swapoff -a"]
    run_cmds = ["touch /root/test"]
  }
}
`
	// Creating configuration
	fullConfig := providerConfigTemplate + resourceConfigTemplate
	fullConfigUpdate := providerConfigTemplate + resourceConfigUpdateTemplate
	//fullConfigImport := providerConfigTemplate + resourceConfigImportTemplate

	// Execute the test.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		// Define the steps of the test.
		Steps: []resource.TestStep{
			{
				// **Create Step**
				// Terraform configuration for creating the VM.
				Config: fullConfig,

				// Check function to verify the resource state after creation.
				Check: resource.ComposeTestCheckFunc(
					// Verify that the VM ID is set.
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "id"),

					// Top-Level Attributes
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "name", "test-vm2"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "cpus", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "ram", "2048"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "vdc_id"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "pool_selector"),

					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "description", "This is a test VM for testing purposes."),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "cpu_priority", "10"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "boot_media", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "vcpu_class", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "os_type", "6"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "os_profile", "4001"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "node"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "action", "start"),

					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "status", "3"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "admin_status", "3"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "oper_status", "3"),

					// Nested Attributes: Guest
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_used"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_balloon_performed"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_balloon_requested"),

					// Nested Attributes: Disks
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.#", "1"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "disks.0.guid"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.size", "20"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.slot", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.iops_limit", "200"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.mbps_limit", "256"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.label", "Primary Disk"),

					// Nested Attributes: Disks -> Sector Size
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.sector_size.logical", "512"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.sector_size.physical", "4096"),
				),
			},
			{
				// **Update Step**
				// Terraform configuration with updated attributes.
				Config: fullConfigUpdate,

				// Check function to verify the resource state after the update.
				Check: resource.ComposeTestCheckFunc(
					// Verify that the VM ID is set.
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "id"),

					// Top-Level Attributes
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "name", "test-vm2"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "cpus", "2"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "ram", "4096"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "vdc_id"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "pool_selector"),

					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "description", "This is a test VM for testing purposes."),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "cpu_priority", "10"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "boot_media", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "vcpu_class", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "os_type", "6"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "os_profile", "4001"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "node"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "action", "stop"),

					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "status", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "admin_status", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "oper_status", "1"),

					// Nested Attributes: Guest
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_used"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_balloon_performed"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "guest.ram_balloon_requested"),

					// Nested Attributes: Disks
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.#", "2"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "disks.0.guid"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.size", "25"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.slot", "1"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.iops_limit", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.mbps_limit", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.label", "Primary Disk-1"),
					// Nested Attributes: Disks -> Sector Size
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.sector_size.logical", "512"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.0.sector_size.physical", "4096"),

					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "disks.#"),
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm2", "disks.1.guid"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.size", "20"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.slot", "2"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.iops_limit", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.mbps_limit", "0"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.label", "Secondary Disk-2"),

					// Nested Attributes: Disks -> Sector Size
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.sector_size.logical", "512"),
					resource.TestCheckResourceAttr("vstack_vm.test_vm2", "disks.1.sector_size.physical", "4096"),
				),
			},
			{
				// **Import Step**
				Config: fullConfig,

				// Import State flags.
				ResourceName: "vstack_vm.test_vm2",
				ImportState:  true,
				// Добавляем ImportStateVerifyIgnore для игнорирования определённых атрибутов при импорте.
				ImportStateVerifyIgnore: []string{
					"guest.ram_used",
					"guest.resolver.%",
					"guest.users.%",
					"guest.users.root.%",
					"guest.users.user.%",
					"guest.boot_cmds.#",
					"guest.boot_cmds.0",
					"guest.hostname",
					"guest.resolver.#",
					"guest.resolver.name_server.#",
					"guest.resolver.name_server.0",
					"guest.resolver.name_server.1",
					"guest.resolver.name_server.2",
					"guest.resolver.search",
					"guest.run_cmds.#",
					"guest.run_cmds.0",
					"guest.ssh_password_auth",
					"guest.users.#",
					"guest.users.root.#",
					"guest.users.root.password",
					"guest.users.root.ssh_authorized_keys.#",
					"guest.users.root.ssh_authorized_keys.0",
					"guest.users.root.ssh_authorized_keys.1",
					"guest.users.user.#",
					"guest.users.user.password",
					"guest.users.user.ssh_authorized_keys.#",
					"guest.users.user.ssh_authorized_keys.0",
					"guest.users.user.ssh_authorized_keys.1",
				},
				ImportStateId:     "",
				ImportStateVerify: true,
			},
		},
	})
}
