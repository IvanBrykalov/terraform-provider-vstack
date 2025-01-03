// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package helper

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// MapDisksToModel maps a slice of vstack_api.Disk structs to a slice of models.DiskModel structs.
// It performs validation on each disk's fields and converts data types as necessary.
//
// Parameters:
// - disks: A slice of vstack_api.Disk structs representing disks fetched from the API.
//
// Returns:
// - A slice of models.DiskModel structs ready to be used in the Terraform resource.
// - An error if any validation fails or if there is an issue during the mapping process.
func MapDisksToModel(disks []vstack_api.Disk) ([]models.DiskModel, error) {
	// Initialize the result slice with the same length as the input disks slice.
	result := make([]models.DiskModel, len(disks))

	// Iterate over each disk to validate and map its fields.
	for i, disk := range disks {
		// Validation for required string fields.
		if err := validateString(disk.GUID, fmt.Sprintf("Disk[%d].GUID", i)); err != nil {
			return nil, err
		}
		if err := validateString(disk.Label, fmt.Sprintf("Disk[%d].Label", i)); err != nil {
			return nil, err
		}

		// Validation for required integer fields.
		if err := validateInt64(disk.Size, fmt.Sprintf("Disk[%d].Size", i)); err != nil {
			return nil, err
		}
		if err := validateInt64(disk.Slot, fmt.Sprintf("Disk[%d].Slot", i)); err != nil {
			return nil, err
		}

		// Validation for optional integer fields.
		if disk.IOPSLimit != nil {
			if err := validateInt64(*disk.IOPSLimit, fmt.Sprintf("Disk[%d].IOPSLimit", i)); err != nil {
				return nil, err
			}
		}
		if disk.MBPSLimit != nil {
			if err := validateInt64(*disk.MBPSLimit, fmt.Sprintf("Disk[%d].MBPSLimit", i)); err != nil {
				return nil, err
			}
		}

		// Handle the SectorSize field, which is optional.
		var sectorSize types.Object
		if disk.SectorSize != nil {
			// Create a Terraform ObjectValue for SectorSizeModel.
			sectorSizeVal, diags := types.ObjectValue(map[string]attr.Type{
				"logical":  types.Int64Type,
				"physical": types.Int64Type,
			}, map[string]attr.Value{
				"logical":  types.Int64Value(disk.SectorSize.Logical),
				"physical": types.Int64Value(disk.SectorSize.Physical),
			})
			if diags.HasError() {
				// Collect diagnostic errors into a slice of strings.
				var diagErrors []string
				for _, d := range diags {
					diagErrors = append(diagErrors, d.Detail())
				}
				return nil, fmt.Errorf("failed to create SectorSize object for Disk[%d]: %s", i, diagErrors)
			}
			sectorSize = sectorSizeVal
		} else {
			// If SectorSize is nil, set it to a null Terraform object with the expected schema.
			sectorSize = types.ObjectNull(map[string]attr.Type{
				"logical":  types.Int64Type,
				"physical": types.Int64Type,
			})
		}

		// Map the validated and converted fields to the DiskModel struct.
		result[i] = models.DiskModel{
			GUID:       types.StringValue(disk.GUID),
			Size:       types.Int64Value(ConvertBytesToGb(disk.Size)),
			Slot:       types.Int64Value(disk.Slot),
			Label:      types.StringValue(disk.Label),
			IopsLimit:  int64NullIfNil(disk.IOPSLimit),
			MbpsLimit:  int64NullIfNil(disk.MBPSLimit),
			SectorSize: sectorSize,
		}
	}

	return result, nil
}
