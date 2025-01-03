// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package helper

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// MapRespToState maps the API response (VmGetResult) to the Terraform state (VMResourceModel).
// It performs validation on each field and converts data types as necessary.
//
// Parameters:
// - resp: The API response of type vstack_api.VmGetResult containing VM data.
// - state: The current Terraform state of type models.VMResourceModel.
//
// Returns:
// - An updated models.VMResourceModel reflecting the API response.
// - An error if any validation fails during the mapping process.
func MapRespToState(resp vstack_api.VmGetResult, state models.VMResourceModel) (models.VMResourceModel, error) {
	// Validate required integer fields.
	if err := validateInt64(resp.Data.ID, "ID"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.CPUs, "CPUs"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.RAM, "RAM"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.CpuPriority, "CPU Priority"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.BootMediaID, "Boot Media ID"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.VcpuClass, "Vcpu Class"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.OsType, "OS Type"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.Vdc, "Vdc ID"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.AdminStatus, "Admin Status"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.Node, "Node"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.CreateCompleted, "Create Completed"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.Locked, "Locked"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.Status, "Status"); err != nil {
		return state, err
	}
	if err := validateInt64(resp.Data.OperStatus, "Oper Status"); err != nil {
		return state, err
	}

	// Validate required string fields.
	if err := validateString(resp.Data.Name, "Name"); err != nil {
		return state, err
	}

	if err := validateString(resp.Data.OsProfile, "OS Profile"); err != nil {
		return state, err
	}
	if err := validateString(fmt.Sprintf("%v", resp.Data.RootDataset), "Root Dataset"); err != nil {
		return state, err
	}
	if err := validateString(resp.Data.RootDatasetName, "Root Dataset Name"); err != nil {
		return state, err
	}
	if err := validateString(resp.Data.Pool, "Pool VM resident"); err != nil {
		return state, err
	}
	if err := validateString(resp.Data.UEFI, "UEFI"); err != nil {
		return state, err
	}

	// Map the validated and converted fields to the VMResourceModel struct.
	state.ID = types.Int64Value(resp.Data.ID)
	state.Name = types.StringValue(resp.Data.Name)
	if resp.Data.Description != nil {
		state.Description = types.StringValue(*resp.Data.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.CPUs = types.Int64Value(resp.Data.CPUs)
	state.RAM = types.Int64Value(ConvertBytesToMb(resp.Data.RAM))
	state.CPUPriority = types.Int64Value(resp.Data.CpuPriority)
	state.BootMedia = types.Int64Value(resp.Data.BootMediaID)
	state.VcpuClass = types.Int64Value(resp.Data.VcpuClass)
	state.OsType = types.Int64Value(resp.Data.OsType)
	state.OsProfile = types.StringValue(resp.Data.OsProfile)
	state.VdcID = types.Int64Value(resp.Data.Vdc)
	state.AdminStatus = types.Int64Value(resp.Data.AdminStatus)
	state.Node = types.Int64Value(resp.Data.Node)
	state.Uefi = types.StringValue(resp.Data.UEFI)
	state.CreateCompleted = types.Int64Value(resp.Data.CreateCompleted)
	state.Locked = types.Int64Value(resp.Data.Locked)
	state.RootDataset = types.StringValue(fmt.Sprintf("%v", resp.Data.RootDataset))
	state.RootDatasetName = types.StringValue(resp.Data.RootDatasetName)
	state.PoolSelector = types.StringValue(resp.Data.Pool)
	state.Status = types.Int64Value(resp.Data.Status)
	state.OperStatus = types.Int64Value(resp.Data.OperStatus)
	state.Action = types.StringValue(getActionFromStatus(resp.Data.OperStatus))

	// Handle Guest (nullable)
	if resp.Data.Guest != nil {
		guestModel := &models.GuestModel{
			RamUsed:            types.Int64Value(resp.Data.Guest.RamUsed),
			RamBallonPerformed: types.Int64Value(resp.Data.Guest.RamBallonPerformed),
			RamBallonRequested: types.Int64Value(resp.Data.Guest.RamBallonRequested),
		}

		// Check if Hostname was set in the configuration
		if state.Guest != nil && !state.Guest.Hostname.IsNull() && !state.Guest.Hostname.IsUnknown() {
			guestModel.Hostname = state.Guest.Hostname
		} else {
			guestModel.Hostname = types.StringNull()
		}

		// Check if BootCmds were set in the configuration
		if state.Guest != nil && !state.Guest.BootCmds.IsNull() && !state.Guest.BootCmds.IsUnknown() {
			guestModel.BootCmds = state.Guest.BootCmds
		} else {
			guestModel.BootCmds = types.ListNull(types.StringType)
		}

		// Check if RunCmds were set in the configuration
		if state.Guest != nil && !state.Guest.RunCmds.IsNull() && !state.Guest.RunCmds.IsUnknown() {
			guestModel.RunCmds = state.Guest.RunCmds
		} else {
			guestModel.RunCmds = types.ListNull(types.StringType)
		}

		// Check if Resolver was set in the configuration
		if state.Guest != nil && state.Guest.Resolver != nil {
			guestModel.Resolver = state.Guest.Resolver
		} else {
			guestModel.Resolver = nil
		}

		// Check if Users were set in the configuration
		if state.Guest != nil && len(state.Guest.Users) > 0 {
			guestModel.Users = state.Guest.Users
		} else {
			guestModel.Users = make(map[string]models.UserModel)
		}

		// Check if SSHPasswordAuth was set in the configuration
		if state.Guest != nil && !state.Guest.SSHPasswordAuth.IsNull() && !state.Guest.SSHPasswordAuth.IsUnknown() {
			guestModel.SSHPasswordAuth = state.Guest.SSHPasswordAuth
		} else {
			// Set default value if it is not present
			guestModel.SSHPasswordAuth = types.Int64Value(0)
		}

		state.Guest = guestModel
	} else {
		state.Guest = nil // If guest is missing
	}

	// Map the disks from the API response to the Terraform state.
	if disks, err := MapDisksToModel(resp.Data.Disks); err != nil {
		return state, err
	} else {
		state.Disks = disks
	}

	return state, nil
}

// getActionFromStatus determines the appropriate action based on the VM's operational status.
// It returns "start" if the VM is started, "stop" if the VM is offline or created, and an empty string otherwise.
//
// Parameters:
// - status: The operational status code of the VM.
//
// Returns:
// - A string representing the action to be performed.
func getActionFromStatus(status int64) string {
	switch status {
	case Status.Started:
		return "start"
	case Status.Offline, Status.Created:
		return "stop"
	default:
		return ""
	}
}
