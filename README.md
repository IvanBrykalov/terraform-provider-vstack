# Terraform Provider Vstack

## Overview
The Terraform vStack Provider allows you to manage vStack resources using HashiCorp Terraform. This provider enables you to create, update, and delete virtual machines (VMs) and network interface cards (NICs) within your vStack environment seamlessly.

## Features
* VM Management: Create, update, and delete virtual machines with customizable configurations.
* NIC Management: Attach and manage network interface cards to your VMs.
* Import Functionality: Import existing vStack resources into your Terraform state for management.
* Comprehensive Testing: Includes acceptance tests to ensure resource integrity and provider reliability.

## Prerequisites
* Terraform: Version 1.3.0 or higher.
* Go: Version 1.16 or higher (required for building the provider).
* vStack Account: Access credentials for your vStack environment.

## Installation
To use the released version of the Terraform provider in your environment, run `terraform init` and Terraform will automatically install the provider from the Terraform registry.

## Upgrade provider version
The provider is not updated automatically. After each new release, you can run the following command to update the vendor:
```
terraform init -upgrade
```

## License

This project is licensed under the **MIT License**.

## Support
If you encounter any issues or have questions, please open an issue on the GitHub Issues page.
