// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package vstack_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// BaseJSONRPCResponse contains the common fields for all JSON-RPC responses.
// It serves as a base structure for unpacking the generic parts of the response.
type BaseJSONRPCResponse struct {
	ID      string          `json:"id"`               // The ID of the JSON-RPC request.
	JsonRPC string          `json:"jsonrpc"`          // The JSON-RPC protocol version.
	Error   *Error          `json:"error,omitempty"`  // Contains error information if the request failed.
	Result  json.RawMessage `json:"result,omitempty"` // The raw JSON result to be parsed separately.
}

// Error represents the structure for API errors.
// It implements the error interface for seamless error handling.
type Error struct {
	Code    int    `json:"code"`    // The error code returned by the API.
	Message string `json:"message"` // A descriptive error message.
}

// Error implements the error interface for the Error struct.
// It returns a formatted error string.
func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// DoRequest sends a JSON-RPC request, unpacks the basic response structure,
// and parses the "result" field into the provided resultContainer if needed.
//
// Parameters:
// - requestPayload: The JSON-RPC request payload as a map.
// - authCookie: The authentication cookie for the request.
// - baseURL: The base URL of the API endpoint.
// - client: The HTTP client used to send the request.
// - resultContainer: A pointer to the struct where the "result" field will be unmarshaled.
//
// Returns:
//   - error: An error object if the request fails, the response contains an error,
//     or the "result" field cannot be decoded.
func DoRequest(
	requestPayload map[string]interface{},
	authCookie string,
	baseURL string,
	client *http.Client,
	resultContainer interface{}, // Pointer to the struct where "result" will be parsed.
) error {

	// 1. Serialize the request payload to JSON.
	reqBody, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("DoRequest: error marshaling payload: %w", err)
	}

	// 2. Create a new HTTP POST request.
	apiReq, err := http.NewRequest("POST", baseURL+"/.api/V4/.req/", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("DoRequest: error creating request: %w", err)
	}
	apiReq.Header.Set("Content-Type", "application/json")            // Set the content type to JSON.
	apiReq.Header.Set("X-Session-Auth", "APIEndpoint00="+authCookie) // Set the authentication header.

	// 3. Execute the HTTP request.
	apiResp, err := client.Do(apiReq)
	if err != nil {
		return fmt.Errorf("DoRequest: HTTP error: %w", err)
	}
	defer func() {
		if err := apiResp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// 4. Decode the basic JSON-RPC response structure.
	var baseResp BaseJSONRPCResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&baseResp); err != nil {
		return fmt.Errorf("DoRequest: decode error: %w", err)
	}

	// 5. Check if the response contains an error.
	if baseResp.Error != nil {
		return fmt.Errorf("DoRequest: API error: %s", baseResp.Error.Error())
	}

	// 6. If a resultContainer is provided, unmarshal the "result" field into it.
	if resultContainer != nil {
		if err := json.Unmarshal(baseResp.Result, resultContainer); err != nil {
			return fmt.Errorf("DoRequest: error decoding result field: %w", err)
		}
	}

	return nil
}
