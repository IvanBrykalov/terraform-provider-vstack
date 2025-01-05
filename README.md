# Terraform Provider vStack
*A Terraform provider for managing resources within a [vStack](https://vstack.com/) environment.*

## Overview

The Terraform vStack Provider allows you to manage vStack cloud resources using HashiCorp Terraform. This provider enables you to create, update, and delete virtual machines (VMs) and network interface cards (NICs) within your vStack environment seamlessly.

## Features
* VM Management: Create, update, and delete virtual machines with customizable configurations.
* NIC Management: Attach and manage network interface cards to your VMs.
* Import Functionality: Import existing vStack resources into your Terraform state for management.
* Comprehensive Testing: Includes acceptance tests to ensure resource integrity and provider reliability.

## Prerequisites
* Terraform: Version 1.6.0 or higher.
* Go: Version 1.22.7 or higher (required for building the provider).
* vStack Account: Access credentials for your vStack environment.

## Installation
To use the released version of the Terraform provider in your environment, run `terraform init` and Terraform will automatically install the provider from the Terraform registry.

## Upgrade provider version
**Highly recommend using the latest version of the provider**

The provider is not updated automatically. After each new release, you can run the following command to update the vendor:
```
terraform init -upgrade
```

## Example usage

Provider Configuration
```
terraform {
  required_providers {
    vstack = {
      source  = "IvanBrykalov/vstack"
    }
  }
}

provider "vstack" {
  host     = "https://vstack.example.com"
  username = "user"
  password = "password"
}
```
The following is a basic example of creating a VM resource using all the VM parameters and setting the values directly.
```
# Manage VM instance
resource "vstack_vm" "example" {
  name          = "example"
  description   = "This is a test VM for demonstration purposes."
  cpus          = 4
  ram           = 4096
  cpu_priority  = 10
  boot_media    = 0
  vcpu_class    = 1
  os_type       = 6
  os_profile    = "4001"                  #This is string value
  vdc_id        = 1234
  pool_selector = "12345678911234567891"  #This is string value

  action = "start"

  disks = [
    {
      size       = 40
      slot       = 1
      label      = "Primary Disk"
      iops_limit = 0
      mbps_limit = 0
    }
  ]

  guest = {
    hostname          = "example"
    ssh_password_auth = 1
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1..."
        ]
        password = "password"
      }
    }
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "example.local"
    }
    boot_cmds = ["swapoff -a"]
    run_cmds  = ["systemctl restart ntpd"]
  }
}

#Setup VM network interface
resource "vstack_nic" "example_nic1" {
  vm_id           = vstack_vm.example.id
  network_id      = 1234
  slot            = 1
  address         = "192.168.0.2"
  ratelimit_mbits = 0
  depends_on      = [vstack_vm.example]
}
```
Below is a basic example of creating a VM resource with minimum parameters, one disk using an OS profile looked up by name. 
The code demonstrates a “flattened” approach to map profile names directly to IDs.
```
data "vstack_vm_profile" "profiles" {}

# Flatten all profiles into one map: profile_name -> { os_type_id, os_profile_id }
locals {
  all_profiles_map = merge([
    for os_type in data.vstack_vm_profile.profiles.os_types : {
      for p in os_type.profiles : p.name => {
        os_type_id    = os_type.id
        os_profile_id = p.id
      }
    }
  ]...)
}

variable "selected_profile_name" {
  type    = string
  default = "Ubuntu 20.04.6 v2"
}

resource "vstack_vm" "example" {
  name   = "demo-vm"
  cpus   = 2
  ram    = 2048       # in MB
  vdc_id = 1234

  # Auto-lookup the OS type and profile (OS distributive) by name
  os_profile = local.all_profiles_map[var.selected_profile_name].os_profile_id

  disks = [
    {
      size = 20       # in GB
      slot = 1
    }
  ]

  guest = {
    hostname = "demo-vm"
    users = {
      root = {
        ssh_authorized_keys = []
        password = "password"
      }
    }
  }
}

#Setup VM network interface
resource "vstack_nic" "example_nic1" {
  vm_id           = vstack_vm.example.id
  network_id      = 1234
  slot            = 1
  depends_on      = [vstack_vm.example]
}
```
**For more information about provider parameters, resources and data_sources, see the documentation directory `./doc` in this repository and
cloud provider web-site vStack.**

## Development
The `make` utility must be installed

### Lint
The [golangci-lint](https://golangci-lint.run/welcome/install/) utility must be installed 
```
make lint
```

### Build
```
make build
```

### Install
```
make install
```

### Test
The following OS environment variables must be set:
* `TF_VAR_host` used to connect your vStack cloud
* `TF_VAR_username`  used to connect your vStack cloud
* `TF_VAR_password` used to connect your vStack cloud
* `TF_VAR_vdc_id` used to create VM with required vdc_id and check result
* `TF_VAR_network_id` used to create VM network interface (NIC) and check result
* `TF_VAR_ip_address` used to create VM network interface (NIC) and check result
* `TF_VAR_os_profile` used to create VM with required profile and check result
* `TF_VAR_os_pool_selector` used to create a VM in the required pool and check result


then you can run tests
```
make testacc
```
## License

This project is licensed under the **MIT License**.

## Support
If you encounter any issues or have questions, please open an issue on the GitHub Issues page.
