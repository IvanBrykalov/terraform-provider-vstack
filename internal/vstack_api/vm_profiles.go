// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"fmt"
	"net/http"
)

// VmProfilesResult describes the structure of the response for the "vm-profiles" method.
type VmProfilesResult struct {
	Code CodeUnion `json:"code"`
	Data map[string]struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Profiles []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			MinSize     int64  `json:"min_size"`
		} `json:"profiles"`
	} `json:"data"`
}

// VmProfiles sends a JSON-RPC "vm-profiles" request and returns the parsed result.
func VmProfiles(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
) (VmProfilesResult, error) {
	var result VmProfilesResult

	// Send the request and parse the response into the result structure.
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		return VmProfilesResult{}, fmt.Errorf("VmProfiles: %w", err)
	}

	// Check the response code to ensure the request was successful.
	if result.Code.CodeAsInt() != 1 {
		return result, fmt.Errorf("VmProfiles returned code=%s", result.Code.CodeAsString())
	}

	return result, nil
}
