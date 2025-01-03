package helper

import (
	"fmt"
	"net/http"

	"terraform-provider-vstack/internal/vstack_api"
)

// ActionVM defines an action to be performed on a Virtual Machine (VM).
// It encapsulates the operational status and the corresponding API method required to execute the action.
type ActionVM struct {
	OperStatus int64  // Operational status codes, e.g., Status.Started or Status.Offline
	Method     string // API method name, e.g., "vms-restart" or "vms-stop"
}

// Action maps human-readable action names to their corresponding ActionVM configurations.
// This allows for easy retrieval and execution of VM actions based on user input or other logic.
var Action = map[string]ActionVM{
	"start": {
		OperStatus: Status.Started, // Assumes Status.Started is a predefined constant
		Method:     "vms-restart",
	},
	"stop": {
		OperStatus: Status.Offline, // Assumes Status.Offline is a predefined constant
		Method:     "vms-stop",
	},
}

// Execute performs the specified action on the VM identified by vmID.
// It constructs the appropriate JSON-RPC request and invokes the corresponding API method.
//
// Parameters:
// - vmID: The unique identifier of the VM on which the action is to be performed.
// - client: An HTTP client used to make API requests.
// - authCookie: The authentication cookie required for API access.
// - baseURL: The base URL of the API endpoint.
//
// Returns:
// - An error if the API call fails or if the action execution encounters issues.
func (a ActionVM) Execute(vmID int64, client *http.Client, authCookie, baseURL string) error {
	// Construct the JSON-RPC request payload using the specified method and VM ID.
	requestPayload := BuildJSONRPCRequest(a.Method, map[string]interface{}{
		"id": vmID,
	})

	// Execute the API call to perform the action (e.g., restart or stop the VM).
	// The VmsStartStop function is assumed to handle the specific API interaction.
	_, err := vstack_api.VmsStartStop(requestPayload, authCookie, baseURL, client)
	if err != nil {
		// Wrap and return the error with additional context for easier debugging.
		return fmt.Errorf("SetNicRatelimit: error executing action '%s' for VM ID %d: %w", a.Method, vmID, err)
	}

	// Action executed successfully.
	return nil
}
