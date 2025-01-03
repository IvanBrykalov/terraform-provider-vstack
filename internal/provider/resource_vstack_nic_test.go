// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"
)

// TestAccVStackNIC tests the VStack NIC resource, including Create, Update, and Delete steps.
func TestAccVStackNIC(t *testing.T) {
	// Retrieve the network ID from the environment variable.
	networkIdStr := os.Getenv("TF_VAR_network_id")
	if networkIdStr == "" {
		t.Fatal("Error in TestAccVStackNIC test function: Environment variable TF_VAR_network_id must be set")
	}

	// Define the Terraform configuration template for the resource.
	resourceConfigTemplate := `
variable "vdc_id" {
  type     = number
}
variable "network_id" {
  type     = number
}

resource "vstack_vm" "test_vm3" {
  name          = "test-vm3"
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

  # Guest parameters
  guest = {
    hostname          = "test_vm3"
    ssh_password_auth = 1

    # User configuration
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."
        ]
        password = "rootpassword"
      }
    }

    # DNS resolver settings
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "test.local"
    }

    # Boot commands to execute on startup
    boot_cmds = ["swapoff -a"]
    
    # Commands to run after bootstrap
    run_cmds = ["touch /root/test"]
  }
}

resource "vstack_nic" "test_vm3_nic1" {
  vm_id      = vstack_vm.test_vm3.id
  network_id = var.network_id
  slot       = 1
  address    = "192.168.0.100"
  
  # Ensure the NIC is created after the VM
  depends_on = [vstack_vm.test_vm3]
}
`

	// Define the updated Terraform configuration template for the Update step.
	resourceConfigUpdateTemplate := `
variable "vdc_id" {
 type     = number
}
variable "network_id" {
 type     = number
}

resource "vstack_vm" "test_vm3" {
  name          = "test-vm3"
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

  # Guest parameters
  guest = {
    hostname          = "test_vm3"
    ssh_password_auth = 1

    # User configuration
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB38gLSqKWZXd7Cs5tJR...",
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAMQta/tG89R5cD9e..."
        ]
        password = "rootpassword"
      }
    }

    # DNS resolver settings
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "test.local"
    }

    # Boot commands to execute on startup
    boot_cmds = ["swapoff -a"]

    # Commands to run after bootstrap
    run_cmds = ["touch /root/test"]
  }
}

resource "vstack_nic" "test_vm3_nic1" {
 vm_id           = vstack_vm.test_vm3.id
 network_id      = var.network_id
 slot            = 1
 address         = "192.168.0.100"
 ratelimit_mbits = 40
 depends_on      = [vstack_vm.test_vm3]
}

resource "vstack_nic" "test_vm3_nic2" {
 vm_id           = vstack_vm.test_vm3.id
 network_id      = var.network_id
 slot            = 2
 ratelimit_mbits = 100

 depends_on = [vstack_vm.test_vm3]
}
`

	// Combine the provider configuration with the resource configuration.
	fullConfig := providerConfigTemplate + resourceConfigTemplate
	fullConfigUpdate := providerConfigTemplate + resourceConfigUpdateTemplate

	// Execute the test using Terraform's testing framework.
	resource.Test(t, resource.TestCase{
		// Define the provider factories.
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		// Define the steps of the test.
		Steps: []resource.TestStep{
			{
				// **Create Step**
				// Terraform configuration for creating the VM and NIC.
				Config: fullConfig,

				// Check function to verify the resource state after creation.
				Check: resource.ComposeTestCheckFunc(
					// Ensure the VM resource has an ID set.
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm3", "id"),
					// Ensure the NIC resource has a vm_id set.
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic1", "vm_id"),
					// Verify that the NIC's vm_id matches the VM's id.
					resource.TestCheckResourceAttrPair("vstack_nic.test_vm3_nic1", "vm_id", "vstack_vm.test_vm3", "id"),

					// Check specific attributes of the NIC.
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "network_id", networkIdStr),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "slot", "1"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "address", "192.168.0.100"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "ratelimit_mbits", "0"),
					// Ensure the MAC address is set (computed attribute).
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic1", "mac"),
				),
			},
			{
				// **Update Step**
				// Terraform configuration with updated attributes.
				Config: fullConfigUpdate,

				// Check function to verify the resource state after the update.
				Check: resource.ComposeTestCheckFunc(
					// Ensure the VM resource still has an ID set.
					resource.TestCheckResourceAttrSet("vstack_vm.test_vm3", "id"),
					// Ensure the first NIC still has a vm_id set.
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic1", "vm_id"),
					// Verify that the first NIC's vm_id matches the VM's id.
					resource.TestCheckResourceAttrPair("vstack_nic.test_vm3_nic1", "vm_id", "vstack_vm.test_vm3", "id"),
					// Check updated attributes of the first NIC.
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "network_id", networkIdStr),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "slot", "1"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "address", "192.168.0.100"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic1", "ratelimit_mbits", "40"),
					// Ensure the MAC address of the first NIC is still set.
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic1", "mac"),

					// Check attributes of the second NIC.
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic2", "vm_id"),
					resource.TestCheckResourceAttrPair("vstack_nic.test_vm3_nic2", "vm_id", "vstack_vm.test_vm3", "id"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic2", "network_id", networkIdStr),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic2", "slot", "2"),
					// Ensure the address of the second NIC is set (computed attribute).
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic2", "address"),
					resource.TestCheckResourceAttr("vstack_nic.test_vm3_nic2", "ratelimit_mbits", "100"),
					// Ensure the MAC address of the second NIC is set.
					resource.TestCheckResourceAttrSet("vstack_nic.test_vm3_nic2", "mac"),

					// Check the operational status of the VM.
					resource.TestCheckResourceAttr("vstack_vm.test_vm3", "oper_status", "3"),
				),
			},
			{
				// Step 3: Import the NIC resource
				Config:       fullConfigUpdate,
				ResourceName: "vstack_nic.test_vm3_nic1",

				ImportState: true,

				// Find vm id and nic id to import resource with required id format vm_id/id
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					// Retrieve the NIC resource from the state.
					nicRes, ok := s.RootModule().Resources["vstack_nic.test_vm3_nic1"]
					if !ok {
						return "", fmt.Errorf("Resource vstack_nic.test_vm3_nic1 not found in state")
					}

					// Extract vm_id and id from the NIC resource.
					vmID := nicRes.Primary.Attributes["vm_id"]
					nicID := nicRes.Primary.Attributes["id"]
					if vmID == "" {
						return "", fmt.Errorf("vm_id is empty in state for NIC")
					}
					if nicID == "" {
						return "", fmt.Errorf("nicID is empty in state for NIC")
					}

					// Construct the ImportStateId in the format 'vm_id/port_id'.
					return fmt.Sprintf("%s/%s", vmID, nicID), nil
				},
				// Verify that the imported state matches the existing configuration.
				ImportStateVerify: true,
			},
		},
	})
}
