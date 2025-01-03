// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"fmt"
	"net/http"
)

// VmSetResult represents the structure for the "result" field in the response to the "vm-set" method.
type VmSetResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message,omitempty"`
	} `json:"data,omitempty"`
}

// VmSet sends a JSON-RPC "vm-set" request and parses the result.
// It returns a VmSetResult struct containing the response message or an error if the request fails.
//
// Parameters:
// - requestPayload: The JSON-RPC request payload.
// - authCookie: The authentication cookie for the request.
// - baseURL: The base URL of the API endpoint.
// - client: The HTTP client used to send the request.
//
// Returns:
// - VmSetResult: The result containing the response message.
// - error: An error object if the request fails or the response code is unexpected.
func VmSet(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
) (VmSetResult, error) {

	var result VmSetResult

	// Use the universal DoRequest helper function to send the request and parse the response.
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		// The error may be an HTTP, JSON, or API error (text from DoRequest).
		return VmSetResult{}, fmt.Errorf("VmSet: %w", err)
	}

	// Check the response code to ensure the operation was successful.
	if result.Code.CodeAsInt() != 1 {
		return result, fmt.Errorf("VmSet: unexpected code=%s", result.Code.CodeAsString())
	}

	return result, nil
}
