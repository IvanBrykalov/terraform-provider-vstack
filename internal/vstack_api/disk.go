package vstack_api

import (
	"fmt"
	"net/http"
)

// 1. vms-add-disk

// VmsAddDiskResult represents the structure for the "result" field in the response to the "vms-add-disk" method.
type VmsAddDiskResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		GUID       string `json:"guid"`
		SectorSize struct {
			Logical  int64 `json:"logical"`
			Physical int64 `json:"physical"`
		} `json:"sector_size"`
		Size  int64  `json:"size"`
		Label string `json:"label"`
		Slot  int64  `json:"slot"`
	} `json:"data,omitempty"`
}

// VmsAddDisk sends a "vms-add-disk" request and returns the result (GUID, SectorSize, etc.).
func VmsAddDisk(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmsAddDiskResult, error) {
	var result VmsAddDiskResult
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmsAddDiskResult{}, fmt.Errorf("VmsAddDisk: %w", err)
	}
	return result, nil
}

// 2. vms-disk-resize

// VmsDiskResizeResult represents the structure for the "result" field in the response to the "vms-disk-resize" method.
type VmsDiskResizeResult struct {
	Code CodeUnion `json:"code"`
	// Usually, the data field may be empty or contain a message.
	Data struct {
		Message string `json:"message,omitempty"`
	} `json:"data,omitempty"`
}

// VmsDiskResize sends a "vms-disk-resize" request to change the disk size.
func VmsDiskResize(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmsDiskResizeResult, error) {
	var result VmsDiskResizeResult
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmsDiskResizeResult{}, fmt.Errorf("VmsDiskResize: %w", err)
	}
	return result, nil
}

// 3. vm-remove-disk

// VmRemoveDiskResult represents the structure for the "result" field in the response to the "vm-remove-disk" method.
type VmRemoveDiskResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

// VmRemoveDisk removes a disk (by GUID or slot, etc.).
func VmRemoveDisk(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmRemoveDiskResult, error) {
	var result VmRemoveDiskResult
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmRemoveDiskResult{}, fmt.Errorf("VmRemoveDisk: %w", err)
	}
	return result, nil
}

// 4. vm-ratelimit-disk

// VmRatelimitDiskResult represents the structure for the "result" field when calling "vm-ratelimit-disk".
type VmRatelimitDiskResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

// VmRatelimitDisk sends a "vm-ratelimit-disk" request to set IOPS/MBPS limits on a disk.
func VmRatelimitDisk(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmRatelimitDiskResult, error) {
	var result VmRatelimitDiskResult
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmRatelimitDiskResult{}, fmt.Errorf("VmRatelimitDisk: %w", err)
	}
	return result, nil
}

// 5. vm-disk-set-label

// VmDiskSetLabelResult represents the structure for the "result" field when calling "vm-disk-set-label".
type VmDiskSetLabelResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		// The API might return some info.
		Message string `json:"message,omitempty"`
	} `json:"data,omitempty"`
}

// VmDiskSetLabel sends a "vm-disk-set-label" request to change the label of a disk.
func VmDiskSetLabel(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmDiskSetLabelResult, error) {
	var result VmDiskSetLabelResult
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmDiskSetLabelResult{}, fmt.Errorf("VmDiskSetLabel: %w", err)
	}
	return result, nil
}
