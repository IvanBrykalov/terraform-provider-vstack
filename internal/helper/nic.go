package helper

import (
	"fmt"
	"net/http"

	"terraform-provider-vstack/internal/vstack_api"
)

// FindNicInVmGet searches for the Network Interface Card (NIC) with the specified portID
// within the details of a Virtual Machine (VM) obtained from the API.
//
// Parameters:
// - client: An HTTP client used to make API requests.
// - authCookie: The authentication cookie required for API access.
// - baseURL: The base URL of the API endpoint.
// - vmID: The unique identifier of the VM.
// - portID: The unique identifier of the NIC to be searched.
//
// Returns:
// - A vstack_api.NetworkPort struct representing the found NIC.
// - An error if the NIC is not found or if any API request fails.
func FindNicInVmGet(
	client *http.Client,
	authCookie string,
	baseURL string,
	vmID int64,
	portID int64,
) (vstack_api.NetworkPort, error) {

	// Build the JSON-RPC payload for the "vm-get" method.
	requestPayload := BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": vmID,
	})

	// Execute the "vm-get" API call to retrieve VM details.
	vmResp, err := vstack_api.VmGet(requestPayload, authCookie, baseURL, client)
	if err != nil {
		return vstack_api.NetworkPort{}, fmt.Errorf("FindNicInVmGet: error calling vstack_api.VmGet: %w", err)
	}

	// Extract the list of network ports from the VM response.
	networkPorts := vmResp.Data.NetworkPorts

	// Iterate through the network ports to find the one matching the specified portID.
	for _, p := range networkPorts {
		if p.PortID == portID {
			return p, nil
		}
	}

	// If the NIC with the specified portID is not found, return an error.
	return vstack_api.NetworkPort{}, fmt.Errorf("FindNicInVmGet: NIC with port_id=%d not found in VM (id=%d)", portID, vmID)
}

// SetNicRatelimit updates the rate limit (in megabits) for a specific NIC within a VM.
// It calls the "vm-ratelimit-nic" API method to apply the new rate limit.
//
// Parameters:
// - client: An HTTP client used to make API requests.
// - authCookie: The authentication cookie required for API access.
// - baseURL: The base URL of the API endpoint.
// - vmID: The unique identifier of the VM.
// - portID: The unique identifier of the NIC whose rate limit is to be updated.
// - ratelimitMbits: The new rate limit value in megabits.
//
// Returns:
// - An error if the API call fails or if the rate limit update is unsuccessful.
func SetNicRatelimit(
	client *http.Client,
	authCookie string,
	baseURL string,
	vmID int64,
	portID int64,
	ratelimitMbits int64,
) error {
	// Build the JSON-RPC payload for the "vm-ratelimit-nic" method.
	reqPayload := BuildJSONRPCRequest("vm-ratelimit-nic", map[string]interface{}{
		"vm_id":           vmID,
		"port_id":         portID,
		"ratelimit_mbits": ratelimitMbits,
	})

	// Execute the "vm-ratelimit-nic" API call to update the NIC's rate limit.
	_, err := vstack_api.VmRatelimitNic(reqPayload, authCookie, baseURL, client)
	if err != nil {
		// API returned an error; wrap it with additional context and return.
		return fmt.Errorf("SetNicRatelimit: error from API: %w", err)
	}

	// Successfully updated the NIC's rate limit.
	return nil
}
