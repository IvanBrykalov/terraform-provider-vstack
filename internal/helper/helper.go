// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package helper

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"
	"strings"
	"sync"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// vmLocks stores a *sync.Mutex for each VM by its ID.
var vmLocks sync.Map // map[int64]*sync.Mutex.

func GetVMLock(vmID int64) (*sync.Mutex, error) {
	actual, _ := vmLocks.LoadOrStore(vmID, &sync.Mutex{})
	mutex, ok := actual.(*sync.Mutex)
	if !ok {
		return nil, fmt.Errorf("unexpected type in vmLocks for vmID %d", vmID)
	}
	return mutex, nil
}

// BuildJSONRPCRequest builds a JSON-RPC request with the specified method and parameters.
func BuildJSONRPCRequest(method string, params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":      uuid.NewString(),
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
}

// PerformAction performs the specified action on the VM.
// Returns an error if the operation fails or the action is unsupported.
func PerformAction(client *http.Client, authCookie, baseURL string, vmID int64, actionName string) error {
	actionName = strings.ToLower(actionName)
	action, exists := Action[actionName]
	if !exists {
		return fmt.Errorf("unsupported action: %s", actionName)
	}

	if err := action.Execute(vmID, client, authCookie, baseURL); err != nil {
		return fmt.Errorf("failed to perform action '%s' on VM: %w", actionName, err)
	}

	return nil
}

// VmStatuses defines various operational statuses for VMs.
type VmStatuses struct {
	Started       int64
	Offline       int64
	Starting      int64
	StartFailed   int64
	Stopping      int64
	StopFailed    int64
	Creating      int64
	Deleting      int64
	Deleted       int64
	Created       int64
	Suspended     int64
	Suspending    int64
	SuspendFailed int64
	Resuming      int64
	ResumeFailed  int64
	CreateFailed  int64
	DeleteFailed  int64
}

// Status holds the different statuses a VM can have.
var Status = VmStatuses{
	Offline:       1,
	Starting:      2,
	Started:       3,
	StartFailed:   4,
	Stopping:      5,
	StopFailed:    6,
	Creating:      7,
	Deleting:      8,
	Deleted:       9,
	Created:       10,
	Suspended:     11,
	Suspending:    12,
	SuspendFailed: 13,
	Resuming:      14,
	ResumeFailed:  15,
	CreateFailed:  16,
	DeleteFailed:  17,
}

// CheckIfVMIsRunning checks if the VM is running.
// Returns true if the VM is running (OperStatus == Status.Started), otherwise false.
func CheckIfVMIsRunning(client *http.Client, authCookie, baseURL string, vmID int64) (bool, error) {
	payload := BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": vmID,
	})

	vmResp, err := vstack_api.VmGet(payload, authCookie, baseURL, client)
	if err != nil {
		return false, fmt.Errorf("failed to get VM status: %w", err)
	}

	// Check OperStatus
	if vmResp.Data.OperStatus == Status.Started {
		return true, nil
	}

	return false, nil
}

// int64NullIfNil converts a pointer to int64 into a Terraform types.Int64.
// If the pointer is nil, it returns a null Int64; otherwise, it returns the actual value.
func int64NullIfNil(val *int64) types.Int64 {
	if val == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*val)
}

// validateInt64 verifies that the provided value is of type int64 and is non-negative.
// It returns an error if the validation fails.
func validateInt64(value interface{}, fieldName string) error {
	v, ok := value.(int64)
	if !ok {
		return fmt.Errorf("invalid type for field %s, expected int64 but got %T", fieldName, value)
	}
	if v < 0 {
		return fmt.Errorf("invalid value for field %s, must be non-negative", fieldName)
	}
	return nil
}

// validateString verifies that the provided value is of type string.
// It returns an error if the validation fails.
func validateString(value interface{}, fieldName string) error {
	_, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type for field %s, expected string but got %T", fieldName, value)
	}
	return nil
}

// FormatDisks transforms a slice of DiskModel into a slice of maps suitable for API requests.
// Each disk's properties are converted into key-value pairs expected by the API.
func FormatDisks(disks []models.DiskModel) []map[string]interface{} {
	formatted := []map[string]interface{}{}
	for _, disk := range disks {
		diskMap := map[string]interface{}{
			"size":       ConvertGbToBytes(disk.Size.ValueInt64()), // Convert size from GB to bytes
			"slot":       disk.Slot.ValueInt64(),                   // Slot number for the disk
			"iops_limit": disk.IopsLimit.ValueInt64(),              // IOPS limit for the disk
			"mbps_limit": disk.MbpsLimit.ValueInt64(),              // MBps limit for the disk
			"label":      disk.Label.ValueString(),                 // Label/name for the disk
		}

		formatted = append(formatted, diskMap)
	}
	return formatted
}

// ApplyDefaultSectorSize ensures that a disk has default sector sizes if they are not already set.
// If SectorSize is null, it sets both logical and physical sector sizes to 512 bytes.
// If SectorSize exists but either logical or physical is null, it sets the missing values to 512 bytes.
func ApplyDefaultSectorSize(disk *models.DiskModel) {
	if disk.SectorSize.IsNull() {
		sectorSize, _ := types.ObjectValue(
			map[string]attr.Type{
				"logical":  types.Int64Type, // Logical sector size
				"physical": types.Int64Type, // Physical sector size
			},
			map[string]attr.Value{
				"logical":  types.Int64Value(512), // Default logical sector size
				"physical": types.Int64Value(512), // Default physical sector size
			},
		)
		disk.SectorSize = sectorSize
	} else {
		attributes := disk.SectorSize.Attributes()
		if logical, ok := attributes["logical"].(types.Int64); ok && logical.IsNull() {
			attributes["logical"] = types.Int64Value(512) // Set default logical sector size
		}
		if physical, ok := attributes["physical"].(types.Int64); ok && physical.IsNull() {
			attributes["physical"] = types.Int64Value(512) // Set default physical sector size
		}

		updatedSectorSize, _ := types.ObjectValue(
			map[string]attr.Type{
				"logical":  types.Int64Type,
				"physical": types.Int64Type,
			},
			attributes,
		)
		disk.SectorSize = updatedSectorSize
	}
}

// ConvertMbToBytes converts megabytes to bytes.
// This is useful for API requests that require byte units.
func ConvertMbToBytes(megaBytesNum int64) int64 {
	bytesNum := megaBytesNum * 1024 * 1024
	return bytesNum
}

// ConvertBytesToMb converts bytes to megabytes.
// This can be used to interpret byte values returned by the API.
func ConvertBytesToMb(bytesNum int64) int64 {
	megaBytesNum := bytesNum / 1024 / 1024
	return megaBytesNum
}

// ConvertGbToBytes converts gigabytes to bytes.
// This is useful for API requests that require byte units.
func ConvertGbToBytes(gigaBytesNum int64) int64 {
	bytesNum := gigaBytesNum * 1024 * 1024 * 1024
	return bytesNum
}

// ConvertBytesToGb converts bytes to gigabytes.
// This can be used to interpret byte values returned by the API.
func ConvertBytesToGb(bytesNum int64) int64 {
	gigaBytesNum := bytesNum / 1024 / 1024 / 1024
	return gigaBytesNum
}
