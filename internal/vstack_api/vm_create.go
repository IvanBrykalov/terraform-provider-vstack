// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"fmt"
	"net/http"
)

// VmCreateResult represents the structure for the "result" field in the response to the "vm-create" method.
type VmCreateResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		AdminStatus     int64               `json:"admin_status"`
		BootMediaID     int64               `json:"boot_media_id"`
		CpuPriority     int64               `json:"cpu_priority"`
		CPUs            int64               `json:"cpus"`
		CreateCompleted int64               `json:"create_completed"`
		Created         int64               `json:"created"`
		Description     *string             `json:"description"`
		Disks           []DiskCreate        `json:"disks"`
		GCID            int64               `json:"gc_id"`
		GCName          string              `json:"gc_name"`
		HVFaults        HVFaultsCreate      `json:"hv_faults"`
		HwVersion       int64               `json:"hw_version"`
		ID              int64               `json:"id"`
		Incarnation     int64               `json:"incarnation"`
		Locked          int64               `json:"locked"`
		Modified        int64               `json:"modified"`
		Name            string              `json:"name"`
		NdmpAddress     string              `json:"ndmp_address"`
		NetworkPorts    []NetworkPortCreate `json:"network_ports"`
		Node            int64               `json:"node"`
		OperStatus      int64               `json:"oper_status"`
		OperStatusTS    int64               `json:"oper_status_ts"`
		OsProfile       string              `json:"os_profile"`
		OsType          int64               `json:"os_type"`
		Pool            string              `json:"pool"`
		RAM             int64               `json:"ram"`
		RootDataset     interface{}         `json:"root_dataset"`
		RootDatasetName string              `json:"root_dataset_name"`
		RStart          int64               `json:"rstart"`
		Status          int64               `json:"status"`
		UEFI            string              `json:"uefi"`
		VcpuClass       int64               `json:"vcpu_class"`
		Vdc             int64               `json:"vdc"`
		Guest           GuestCreate         `json:"guest"`
	} `json:"data"`
}

// DiskCreate describes a disk in the virtual machine.
type DiskCreate struct {
	GUID       string      `json:"guid"`
	Size       int64       `json:"size"`
	Slot       int64       `json:"slot"`
	IOPSLimit  *int64      `json:"iops_limit,omitempty"`
	MBPSLimit  *int64      `json:"mbps_limit,omitempty"`
	Label      string      `json:"label,omitempty"`
	SectorSize *SectorSize `json:"sector_size"`
}

// SectorSizeCreate describes the logical and physical sector sizes of a disk.
type SectorSizeCreate struct {
	Logical  int64 `json:"logical"`
	Physical int64 `json:"physical"`
}

// NetworkPortCreate describes a network port of the virtual machine.
type NetworkPortCreate struct {
	Address        string `json:"address"`
	IPGuard        int64  `json:"ip_guard"`
	MAC            string `json:"mac"`
	NetworkID      int64  `json:"network_id"`
	PortID         int64  `json:"port_id"`
	RatelimitMBits *int64 `json:"ratelimit_mbits"`
	Slot           int64  `json:"slot"`
}

// GuestCreate describes the guest configuration of the virtual machine.
type GuestCreate struct {
	RamUsed *int64 `json:"ram_used"`
}

// HVFaultsCreate describes hardware faults related to the virtual machine.
type HVFaultsCreate struct {
	Sigabort SigabortCreate `json:"sigabort"`
}

// SigabortCreate describes the SIGABRT fault configuration.
type SigabortCreate struct {
	Interval int64 `json:"interval"`
	Restarts int64 `json:"restarts"`
}

// VmCreate sends a JSON-RPC "vm-create" request and parses the result.
func VmCreate(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
) (VmCreateResult, error) {

	var result VmCreateResult

	// Call DoRequest to send the request and parse the response into result.
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmCreateResult{}, fmt.Errorf("VmCreate: %w", err)
	}

	// Check the response code to ensure the VM was created successfully.
	if result.Code.CodeAsInt() != 1 {
		return result, fmt.Errorf("VmCreate returned code=%s", result.Code.CodeAsString())
	}

	return result, nil
}
