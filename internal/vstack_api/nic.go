// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"fmt"
	"net/http"
)

// 1. vms-add-nic

// VmsAddNicResult represents the structure for the "result" field in the response to the "vms-add-nic" method.
type VmsAddNicResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Address        string `json:"address"`
		MAC            string `json:"mac"`
		NetworkID      int64  `json:"network_id"`
		PortID         int64  `json:"port_id"`
		Slot           int64  `json:"slot"`
		IPGuard        int64  `json:"ip_guard"`
		RatelimitMBits *int64 `json:"ratelimit_mbits"`
	} `json:"data,omitempty"`
}

// VmsAddNic sends a JSON-RPC "vms-add-nic" request and parses the result.
func VmsAddNic(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmsAddNicResult, error) {
	var result VmsAddNicResult

	// Use our DoRequest
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmsAddNicResult{}, fmt.Errorf("VmsAddNic: DoRequest error: %w", err)
	}

	return result, nil
}

// 2. vm-remove-nic

// VmRemoveNicResult represents the structure for the "result" field in the response to the "vm-remove-nic" method.
type VmRemoveNicResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

// VmRemoveNic removes a NIC.
func VmRemoveNic(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmRemoveNicResult, error) {
	var result VmRemoveNicResult

	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmRemoveNicResult{}, fmt.Errorf("VmRemoveNic: %w", err)
	}

	return result, nil
}

// 3. vm-ratelimit-nic

// VmRatelimitNicResult represents the structure for the "result" field when calling "vm-ratelimit-nic".
type VmRatelimitNicResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

// VmRatelimitNic sets the ratelimit on a NIC.
func VmRatelimitNic(requestPayload map[string]interface{}, authCookie string, baseURL string, client *http.Client) (VmRatelimitNicResult, error) {
	var result VmRatelimitNicResult

	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmRatelimitNicResult{}, fmt.Errorf("VmRatelimitNic: %w", err)
	}

	return result, nil
}
