// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VMResourceModel represents the schema for a Virtual Machine resource in Terraform.
type VMResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`                // Unique identifier for the VM.
	Name            types.String `tfsdk:"name"`              // Name of the VM.
	Description     types.String `tfsdk:"description"`       // Description of the VM.
	CPUs            types.Int64  `tfsdk:"cpus"`              // Number of CPUs allocated to the VM.
	RAM             types.Int64  `tfsdk:"ram"`               // Amount of RAM (in MB) allocated to the VM.
	CPUPriority     types.Int64  `tfsdk:"cpu_priority"`      // Priority level for CPU allocation.
	BootMedia       types.Int64  `tfsdk:"boot_media"`        // ID of the boot media attached to the VM.
	VcpuClass       types.Int64  `tfsdk:"vcpu_class"`        // Virtual CPU class/type.
	OsType          types.Int64  `tfsdk:"os_type"`           // Operating system type identifier.
	OsProfile       types.String `tfsdk:"os_profile"`        // Profile/configuration for the OS.
	VdcID           types.Int64  `tfsdk:"vdc_id"`            // Identifier for the Virtual Data Center.
	AdminStatus     types.Int64  `tfsdk:"admin_status"`      // Administrative status of the VM.
	Node            types.Int64  `tfsdk:"node"`              // Node identifier where the VM is hosted.
	Uefi            types.String `tfsdk:"uefi"`              // UEFI configuration or status.
	CreateCompleted types.Int64  `tfsdk:"create_completed"`  // Flag indicating if VM creation is completed.
	Locked          types.Int64  `tfsdk:"locked"`            // Flag indicating if the VM is locked.
	RootDataset     types.String `tfsdk:"root_dataset"`      // Root dataset associated with the VM.
	RootDatasetName types.String `tfsdk:"root_dataset_name"` // Name of the root dataset.
	PoolSelector    types.String `tfsdk:"pool_selector"`     // Selector for resource pool allocation.
	Status          types.Int64  `tfsdk:"status"`            // Current status of the VM.
	OperStatus      types.Int64  `tfsdk:"oper_status"`       // Operational status of the VM.
	Action          types.String `tfsdk:"action"`            // Action to be performed on the VM (e.g., start, stop).
	Guest           *GuestModel  `tfsdk:"guest"`             // Guest OS configuration and settings.
	Disks           []DiskModel  `tfsdk:"disks"`             // List of disks attached to the VM.
}

// DiskModel describes a disk attached to the virtual machine.
type DiskModel struct {
	GUID       types.String `tfsdk:"guid"`        // Globally Unique Identifier for the disk.
	Slot       types.Int64  `tfsdk:"slot"`        // Slot number where the disk is attached.
	Size       types.Int64  `tfsdk:"size"`        // Size of the disk (in GB).
	IopsLimit  types.Int64  `tfsdk:"iops_limit"`  // IOPS (Input/Output Operations Per Second) limit for the disk.
	MbpsLimit  types.Int64  `tfsdk:"mbps_limit"`  // Bandwidth limit (in Mbps) for the disk.
	Label      types.String `tfsdk:"label"`       // Human-readable label for the disk.
	SectorSize types.Object `tfsdk:"sector_size"` // Sector size configuration for the disk.
}

// SectorSizeModel describes the logical and physical sector sizes of a disk.
type SectorSizeModel struct {
	Logical  types.Int64 `tfsdk:"logical"`  // Logical sector size (in bytes).
	Physical types.Int64 `tfsdk:"physical"` // Physical sector size (in bytes).
}

// NetworkPortModel describes a network port associated with the virtual machine.
type NetworkPortModel struct {
	ID             types.Int64  `tfsdk:"id"`              // Unique identifier for the network port.
	VmID           types.Int64  `tfsdk:"vm_id"`           // Identifier of the VM to which this port belongs.
	MAC            types.String `tfsdk:"mac"`             // MAC address of the network port.
	Address        types.String `tfsdk:"address"`         // IP address assigned to the network port.
	NetworkID      types.Int64  `tfsdk:"network_id"`      // Identifier for the network.
	IpGuard        types.Int64  `tfsdk:"ip_guard"`        // IP guard configuration/status.
	Slot           types.Int64  `tfsdk:"slot"`            // Slot number where the network port is attached.
	RatelimitMbits types.Int64  `tfsdk:"ratelimit_mbits"` // Rate limit (in Mbits) for the network port.
}

// ResolverModel configures DNS resolver settings within the guest OS.
type ResolverModel struct {
	NameServers []types.String `tfsdk:"name_server"` // List of DNS name servers.
	Search      types.String   `tfsdk:"search"`      // DNS search domains.
}

// GuestModel represents the guest OS configuration and settings for the VM.
type GuestModel struct {
	Users              map[string]UserModel `tfsdk:"users"`             // Map of users configured within the guest OS.
	SSHPasswordAuth    types.Int64          `tfsdk:"ssh_password_auth"` // Flag to enable/disable SSH password authentication.
	Resolver           *ResolverModel       `tfsdk:"resolver"`          // DNS resolver configuration.
	BootCmds           types.List           `tfsdk:"boot_cmds"`         // Commands to execute during boot.
	RunCmds            types.List           `tfsdk:"run_cmds"`          // Commands to execute at runtime.
	Hostname           types.String         `tfsdk:"hostname"`          // Hostname of the guest OS.
	RamUsed            types.Int64          `tfsdk:"ram_used"`
	RamBallonPerformed types.Int64          `tfsdk:"ram_balloon_performed"`
	RamBallonRequested types.Int64          `tfsdk:"ram_balloon_requested"`
}

// UserModel represents a user within the guest OS.
type UserModel struct {
	SSHPublicKeys []types.String `tfsdk:"ssh_authorized_keys"` // List of SSH authorized public keys for the user.
	Password      types.String   `tfsdk:"password"`            // Password for the user account.
}
