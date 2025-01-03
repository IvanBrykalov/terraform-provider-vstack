// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"terraform-provider-vstack/internal/helper"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// UpdateDisks manages the synchronization of disk configurations between the Terraform plan and the actual VM state.
// It handles adding new disks, updating existing ones, and removing disks that are no longer present in the plan.
// The function utilizes helper functions for building JSON-RPC requests and handling API interactions.
func (r *VstackVMResource) UpdateDisks(
	ctx context.Context,
	plan *models.VMResourceModel,
	state *models.VMResourceModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// 1. Ensure all disks in the plan have default sector sizes applied if not already set.
	for i := range plan.Disks {
		helper.ApplyDefaultSectorSize(&plan.Disks[i])
	}

	// 2. Create a map of existing disks by slot for efficient lookup during updates/removals.
	stateDisksBySlot := make(map[int64]models.DiskModel)
	for _, disk := range state.Disks {
		slot := disk.Slot.ValueInt64()
		stateDisksBySlot[slot] = disk
	}

	// 3. Track the slots present in the plan to identify disks that need to be removed.
	planDisksSlots := make(map[int64]bool)
	for _, disk := range plan.Disks {
		planDisksSlots[disk.Slot.ValueInt64()] = true
	}

	// 4. Iterate over the plan's disks to determine if they need to be added or updated.
	for _, disk := range plan.Disks {
		slot := disk.Slot.ValueInt64()

		if stateDisk, exists := stateDisksBySlot[slot]; exists {
			// Disk exists in both state and plan: check for updates.
			diskDiags := r.updateExistingDisk(state, disk, stateDisk)
			diags.Append(diskDiags...)
			if diags.HasError() {
				return diags
			}
		} else {
			// Disk exists in plan but not in state: add it.
			diskDiags := r.addNewDisk(state, disk)
			diags.Append(diskDiags...)
			if diags.HasError() {
				return diags
			}
		}
	}

	// 5. Identify and remove disks that exist in state but are absent in the plan.
	removeDiags := r.removeDisksNotInPlan(state, planDisksSlots)
	diags.Append(removeDiags...)
	if diags.HasError() {
		return diags
	}

	return diags
}

// updateExistingDisk handles the update logic for an existing disk.
// It checks for changes in sector size, disk size, rate limits, and labels,
// and performs the necessary API calls to synchronize the state.
func (r *VstackVMResource) updateExistingDisk(
	state *models.VMResourceModel,
	disk models.DiskModel,
	stateDisk models.DiskModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	slot := disk.Slot.ValueInt64()

	// 1. Check for sector size changes.
	if !disk.SectorSize.Equal(stateDisk.SectorSize) {
		diags.AddError(
			"Sector Size Modification Not Allowed",
			fmt.Sprintf(
				"Cannot modify sector_size for disk in slot %d. To change sector_size, remove the disk and add a new one with the desired sector_size.",
				slot,
			),
		)
		return diags
	}

	// 2. Check if disk size needs to be increased.
	if disk.Size.ValueInt64() > stateDisk.Size.ValueInt64() {
		// Utilize the helper's BuildJSONRPCRequest to construct the payload.
		reqPayload := helper.BuildJSONRPCRequest("vms-disk-resize", map[string]interface{}{
			"id":        state.ID.ValueInt64(),
			"disk_guid": stateDisk.GUID.ValueString(),
			"size":      helper.ConvertGbToBytes(disk.Size.ValueInt64()),
		})

		// Execute the disk resize API call.
		_, err := vstack_api.VmsDiskResize(reqPayload, r.AuthCookie, r.BaseURL, r.Client)
		if err != nil {
			diags.AddError("Error resizing disk", err.Error())
			return diags
		}

		log.Printf("Successfully resized disk in slot %d to %d GB", slot, disk.Size.ValueInt64())
	}

	// 3. Check and update rate limits if they have changed.
	if disk.MbpsLimit.ValueInt64() != stateDisk.MbpsLimit.ValueInt64() ||
		disk.IopsLimit.ValueInt64() != stateDisk.IopsLimit.ValueInt64() {
		reqPayload := helper.BuildJSONRPCRequest("vm-ratelimit-disk", map[string]interface{}{
			"vm_id":      state.ID.ValueInt64(),
			"disk_guid":  stateDisk.GUID.ValueString(),
			"mbps_limit": disk.MbpsLimit.ValueInt64(),
			"iops_limit": disk.IopsLimit.ValueInt64(),
		})

		_, err := vstack_api.VmRatelimitDisk(reqPayload, r.AuthCookie, r.BaseURL, r.Client)
		if err != nil {
			diags.AddError("Error updating disk rate limits", err.Error())
			return diags
		}

		log.Printf("Successfully updated rate limits for disk in slot %d", slot)
	}

	// 4. Check and update the disk label if it has changed.
	if disk.Label.ValueString() != stateDisk.Label.ValueString() {
		reqPayload := helper.BuildJSONRPCRequest("vm-disk-set-label", map[string]interface{}{
			"vm_id": state.ID.ValueInt64(),
			"guid":  stateDisk.GUID.ValueString(),
			"label": disk.Label.ValueString(),
		})

		_, err := vstack_api.VmDiskSetLabel(reqPayload, r.AuthCookie, r.BaseURL, r.Client)
		if err != nil {
			diags.AddError("Error updating disk label", err.Error())
			return diags
		}

		log.Printf("Successfully updated label for disk in slot %d to '%s'", slot, disk.Label.ValueString())
	}

	return diags
}

// addNewDisk handles adding a new disk based on the plan configuration.
// It constructs the necessary JSON-RPC request using helper functions and performs the API call.
func (r *VstackVMResource) addNewDisk(
	state *models.VMResourceModel,
	disk models.DiskModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// Prepare sector size attributes, ensuring defaults are applied.
	sectorSizeAttributes := map[string]interface{}{
		"logical":  512,  // Default logical sector size in bytes
		"physical": 4096, // Default physical sector size in bytes
	}
	if !disk.SectorSize.IsNull() {
		attributes := disk.SectorSize.Attributes()
		if logical, ok := attributes["logical"].(types.Int64); ok && !logical.IsNull() {
			sectorSizeAttributes["logical"] = logical.ValueInt64()
		}
		if physical, ok := attributes["physical"].(types.Int64); ok && !physical.IsNull() {
			sectorSizeAttributes["physical"] = physical.ValueInt64()
		}
	}

	// Utilize the helper's BuildJSONRPCRequest to construct the payload.
	reqPayload := helper.BuildJSONRPCRequest("vms-add-disk", map[string]interface{}{
		"vm_id":       state.ID.ValueInt64(),
		"size":        helper.ConvertGbToBytes(disk.Size.ValueInt64()), // Convert size from GB to bytes
		"slot":        disk.Slot.ValueInt64(),
		"label":       disk.Label.ValueString(),
		"sector_size": sectorSizeAttributes,
		"iops_limit":  disk.IopsLimit.ValueInt64(),
		"mbps_limit":  disk.MbpsLimit.ValueInt64(),
	})

	// Execute the disk addition API call.
	_, err := vstack_api.VmsAddDisk(reqPayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		diags.AddError("Error adding disk", err.Error())
		return diags
	}

	log.Printf("Successfully added new disk in slot %d with label '%s'", disk.Slot.ValueInt64(), disk.Label.ValueString())

	return diags
}

// removeDisksNotInPlan identifies disks present in the state but absent in the plan and removes them.
// It ensures that disks no longer defined in the Terraform configuration are deleted from the VM.
func (r *VstackVMResource) removeDisksNotInPlan(
	state *models.VMResourceModel,
	planDisksSlots map[int64]bool,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, stateDisk := range state.Disks {
		slot := stateDisk.Slot.ValueInt64()

		// If the disk's slot is not present in the plan, it should be removed.
		if !planDisksSlots[slot] {
			// 1. Stop the VM if it's currently running to safely remove the disk.
			if state.OperStatus.ValueInt64() != helper.Status.Offline {
				stopReq := helper.BuildJSONRPCRequest("vms-stop", map[string]interface{}{
					"id": state.ID.ValueInt64(),
				})

				_, err := vstack_api.VmsStartStop(stopReq, r.AuthCookie, r.BaseURL, r.Client)
				if err != nil {
					diags.AddError("Error stopping VM before removing disk", err.Error())
					return diags
				}

				log.Printf("Stopped VM ID %d to remove disk in slot %d", state.ID.ValueInt64(), slot)
			}

			// 2. Construct the JSON-RPC request to remove the disk.
			removeReqPayload := helper.BuildJSONRPCRequest("vm-remove-disk", map[string]interface{}{
				"vm_id":     state.ID.ValueInt64(),
				"disk_guid": stateDisk.GUID.ValueString(),
			})

			// 3. Execute the disk removal API call.
			_, err := vstack_api.VmRemoveDisk(removeReqPayload, r.AuthCookie, r.BaseURL, r.Client)
			if err != nil {
				diags.AddError("Error removing disk", err.Error())
				return diags
			}

			log.Printf("Successfully removed disk in slot %d from VM ID %d", slot, state.ID.ValueInt64())

			// 4. Restart the VM if it was previously running.
			if state.OperStatus.ValueInt64() != helper.Status.Offline {
				restartReq := helper.BuildJSONRPCRequest("vms-restart", map[string]interface{}{

					"id": state.ID.ValueInt64(),
				})

				_, err := vstack_api.VmsStartStop(restartReq, r.AuthCookie, r.BaseURL, r.Client)
				if err != nil {
					diags.AddError("Error restarting VM after removing disk", err.Error())
					return diags
				}

				log.Printf("Restarted VM ID %d after removing disk in slot %d", state.ID.ValueInt64(), slot)
			}
		}
	}

	return diags
}
