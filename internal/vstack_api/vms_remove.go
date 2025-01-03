package vstack_api

import (
	"fmt"
	"net/http"
)

// VmRemoveResult represents the structure for the "result" field in the response to the "vm-remove" method.
type VmRemoveResult struct {
	Code CodeUnion `json:"code"`
	Data struct {
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

// VmRemove sends a JSON-RPC "vm-remove" request and parses the result.
// It returns a VmRemoveResult struct containing the response message or an error if the request fails.
//
// Parameters:
// - requestPayload: The JSON-RPC request payload.
// - authCookie: The authentication cookie for the request.
// - baseURL: The base URL of the API endpoint.
// - client: The HTTP client used to send the request.
//
// Returns:
// - VmRemoveResult: The result containing the response message.
// - error: An error object if the request fails or the response code is unexpected.
func VmRemove(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
) (VmRemoveResult, error) {

	var result VmRemoveResult

	// Call our universal DoRequest helper function.
	// It performs an HTTP POST, checks for errors, and parses the "result" into &result.
	if err := DoRequest(requestPayload, authCookie, baseURL, client, &result); err != nil {
		// The error may be an HTTP, JSON, or API error (text from DoRequest).
		return VmRemoveResult{}, fmt.Errorf("VmRemove: %w", err)
	}

	// Check if the response code equals 1, indicating success.
	if result.Code.CodeAsInt() != 1 {
		return result, fmt.Errorf("VmRemove: unexpected code=%s", result.Code.CodeAsString())
	}

	return result, nil
}
