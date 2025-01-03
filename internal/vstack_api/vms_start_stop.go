// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"fmt"
	"net/http"
)

// VmsStartStopResult represents the structure for the "result" field in the response to the "vms-start-stop" method.
type VmsStartStopResult struct {
	Code CodeUnion `json:"code"` // We use CodeUnion to support both int and string types.

	Data struct {
		Message string `json:"message,omitempty"`
		// If the API returns additional fields, add them here.
	} `json:"data,omitempty"`
}

// VmsStartStop sends a JSON-RPC "vms-start-stop" request and parses the result.
// It returns a VmsStartStopResult struct containing the response message or an error if the request fails.
//
// Parameters:
// - requestPayload: The JSON-RPC request payload.
// - authCookie: The authentication cookie for the request.
// - baseURL: The base URL of the API endpoint.
// - client: The HTTP client used to send the request.
//
// Returns:
// - VmsStartStopResult: The result containing the response message.
// - error: An error object if the request fails or the response code is unexpected.
func VmsStartStop(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
) (VmsStartStopResult, error) {

	var result VmsStartStopResult

	// Call the universal DoRequest helper function.
	// It performs an HTTP POST, checks for errors, and parses the "result" into &result.
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		// The error may be an HTTP, JSON, or API error (text from DoRequest).
		return VmsStartStopResult{}, fmt.Errorf("VmsStartStop: %w", err)
	}

	return result, nil
}
