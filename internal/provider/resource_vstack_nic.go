// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net/http"
	"strconv"
	"strings"
	"terraform-provider-vstack/internal/helper"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// VstackNicResource is the resource responsible for managing a single NIC.
type VstackNicResource struct {
	Client     *http.Client
	BaseURL    string
	AuthCookie string
}

func NewVstackNicResource() resource.Resource {
	return &VstackNicResource{}
}

// Schema defines the schema for the NIC resource.
func (r *VstackNicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the NIC (port_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"vm_id": schema.Int64Attribute{
				Description: "ID of the VM where this NIC is attached.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.Int64Attribute{
				Description: "Network ID.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"slot": schema.Int64Attribute{
				Description: "Slot number for the NIC.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Description: "IP address of the NIC. If not provided, it may be auto-assigned by vStack.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				Description: "MAC address of the NIC.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ratelimit_mbits": schema.Int64Attribute{
				Description: "Rate limit in Mbps for the NIC.",
				Optional:    true,
				Computed:    true,
			},
			"ip_guard": schema.Int64Attribute{
				Description: "IP guard setting for the NIC.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VstackNicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nic"
}

// Configure sets up the resource with provider data.
func (r *VstackNicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if pd, ok := req.ProviderData.(*VStackProvider); ok {
		r.Client = pd.client
		r.AuthCookie = pd.authCookie
		r.BaseURL = pd.Host
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			"The provider data was not of the expected *VStackProvider type.",
		)
	}
}

// Create adds a NIC to the VM and ensures the VM is in a stable state.
func (r *VstackNicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan models.NetworkPortModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := plan.VmID.ValueInt64()
	if vmID == 0 {
		resp.Diagnostics.AddError("Invalid VM ID", "VM ID must be greater than zero.")
		return
	}

	// Retrieve the mutex for the VM and lock it
	mu, err := helper.GetVMLock(vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error on GetVMLock", err.Error())
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// 1. Check if the VM was running
	wasRunning, err := helper.CheckIfVMIsRunning(r.Client, r.AuthCookie, r.BaseURL, vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error checking VM status", err.Error())
		return
	}

	// 2. Stop the VM if it was running
	if wasRunning {
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "stop"); err != nil {
			resp.Diagnostics.AddError("Error stopping VM before adding NIC", err.Error())
			return
		}
	}

	// 3. Build the request to add NIC
	params := map[string]interface{}{
		"id":         vmID,
		"network_id": plan.NetworkID.ValueInt64(),
		"slot":       plan.Slot.ValueInt64(),
	}

	if !plan.RatelimitMbits.IsNull() {
		params["ratelimit_mbits"] = plan.RatelimitMbits.ValueInt64()
	}

	if !plan.Address.IsNull() && plan.Address.ValueString() != "" {
		params["address"] = plan.Address.ValueString()
	}

	if !plan.IpGuard.IsNull() && plan.IpGuard.ValueInt64() != 0 {
		params["ip_guard"] = plan.IpGuard.ValueInt64()
	}

	requestCreatePayload := helper.BuildJSONRPCRequest("vms-add-nic", params)

	// 4. Call the API to add NIC
	addResp, addErr := vstack_api.VmsAddNic(requestCreatePayload, r.AuthCookie, r.BaseURL, r.Client)
	if addErr != nil {
		resp.Diagnostics.AddError("Error adding NIC", addErr.Error())
		return
	}

	// 5. Restart the VM if it was running
	if wasRunning {
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
			resp.Diagnostics.AddError("Error restarting VM after adding NIC", err.Error())
			return
		}
	}

	// 6. Update the state with the added NIC details
	plan.ID = types.Int64Value(addResp.Data.PortID)
	plan.MAC = types.StringValue(addResp.Data.MAC)

	if plan.Address.IsNull() || addResp.Data.Address != "" {
		// If the user didn't specify an address, but the API assigned one
		plan.Address = types.StringValue(addResp.Data.Address)
	}

	plan.IpGuard = types.Int64Value(addResp.Data.IPGuard)

	if addResp.Data.RatelimitMBits != nil {
		plan.RatelimitMbits = types.Int64Value(*addResp.Data.RatelimitMBits)
	} else {
		plan.RatelimitMbits = types.Int64Value(0)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log successful NIC addition
	log.Printf("Successfully added NIC with ID %d to VM ID %d", addResp.Data.PortID, vmID)
}

// Read retrieves the current state of the NIC from vStack and updates the Terraform state.
func (r *VstackNicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.NetworkPortModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.VmID.ValueInt64()
	portID := state.ID.ValueInt64()

	if vmID == 0 || portID == 0 {
		resp.Diagnostics.AddError("Invalid IDs", "VM ID and NIC ID must be greater than zero.")
		return
	}

	// Call vm-get to retrieve VM details and find the NIC
	nic, err := helper.FindNicInVmGet(r.Client, r.AuthCookie, r.BaseURL, vmID, portID)
	if err != nil {
		// If NIC not found, remove the resource from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with the retrieved NIC details
	state.Address = types.StringValue(nic.Address)
	state.MAC = types.StringValue(nic.MAC)
	state.IpGuard = types.Int64Value(nic.IPGuard)

	// Safely handle RatelimitMbits which might be nil
	if nic.RatelimitMBits != nil {
		state.RatelimitMbits = types.Int64Value(*nic.RatelimitMBits)
	} else {
		state.RatelimitMbits = types.Int64Null()
	}

	state.Slot = types.Int64Value(nic.Slot)
	state.NetworkID = types.Int64Value(nic.NetworkID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log successful NIC state read
	log.Printf("Successfully read NIC state for NIC ID %d", portID)
}

// Update modifies the NIC parameters as needed.
func (r *VstackNicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve the desired plan and current state
	var plan, state models.NetworkPortModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := plan.VmID.ValueInt64()
	nicID := plan.ID.ValueInt64()

	if vmID == 0 || nicID == 0 {
		resp.Diagnostics.AddError("Invalid IDs", "VM ID and NIC ID must be greater than zero.")
		return
	}

	// Lock the VM to prevent concurrent operations
	mu, err := helper.GetVMLock(vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error on GetVMLock", err.Error())
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// 1. Collect changed NIC parameters (excluding those that require replacement)
	nicParams := make(map[string]interface{})

	// Only handle ratelimit_mbits as per your requirements
	if plan.RatelimitMbits.ValueInt64() != state.RatelimitMbits.ValueInt64() {
		nicParams["ratelimit_mbits"] = plan.RatelimitMbits.ValueInt64()
	}

	// Note: 'address' changes require replacement as per the schema's PlanModifiers
	// So, if 'address' changes, Terraform will handle resource replacement, and no update is needed here

	// 2. Update NIC parameters if there are changes
	if len(nicParams) > 0 {
		// Use the setNicRatelimit helper function
		newRate, exists := nicParams["ratelimit_mbits"].(int64)
		if exists {
			if err := helper.SetNicRatelimit(r.Client, r.AuthCookie, r.BaseURL, vmID, nicID, newRate); err != nil {
				resp.Diagnostics.AddError("Error updating NIC ratelimit", err.Error())
				return
			}
		}
	}

	// 3. Retrieve the full information of the NIC to set the state
	// Using FindNicInVmGet
	nic, err := helper.FindNicInVmGet(r.Client, r.AuthCookie, r.BaseURL, vmID, nicID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving NIC details", err.Error())
		return
	}

	// Map the API response to Terraform state
	state.Address = types.StringValue(nic.Address)
	state.MAC = types.StringValue(nic.MAC)
	state.IpGuard = types.Int64Value(nic.IPGuard)
	state.RatelimitMbits = types.Int64Value(*nic.RatelimitMBits)
	state.Slot = types.Int64Value(nic.Slot)
	state.NetworkID = types.Int64Value(nic.NetworkID)

	// 4. Set the updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log successful NIC update
	log.Printf("Successfully updated NIC with ID %d", nicID)
}

// Delete removes the NIC from the VM and ensures the VM is in a stable state.
func (r *VstackNicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.NetworkPortModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.VmID.ValueInt64()
	portID := state.ID.ValueInt64()

	if vmID == 0 || portID == 0 {
		resp.Diagnostics.AddError("Invalid IDs", "VM ID and NIC ID must be greater than zero.")
		return
	}

	// Lock the VM to prevent concurrent operations
	mu, err := helper.GetVMLock(vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error on GetVMLock", err.Error())
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// 1. Check if the VM was running
	wasRunning, err := helper.CheckIfVMIsRunning(r.Client, r.AuthCookie, r.BaseURL, vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error checking VM status", err.Error())
		return
	}

	// 2. Stop the VM if it was running
	if wasRunning {
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "stop"); err != nil {
			resp.Diagnostics.AddError("Error stopping VM before removing NIC", err.Error())
			return
		}
	}

	// 3. Remove the NIC
	removeReq := helper.BuildJSONRPCRequest("vm-remove-nic", map[string]interface{}{
		"vm_id":   vmID,
		"port_id": portID,
	})

	_, removeErr := vstack_api.VmRemoveNic(removeReq, r.AuthCookie, r.BaseURL, r.Client)
	if removeErr != nil {
		resp.Diagnostics.AddError("Error deleting NIC", removeErr.Error())
		return
	}

	// 4. Restart the VM if it was running
	if wasRunning {
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
			resp.Diagnostics.AddError("Error restarting VM after removing NIC", err.Error())
			return
		}
	}

	// 5. Remove the resource from Terraform state
	resp.State.RemoveResource(ctx)

	// Log successful NIC deletion
	log.Printf("Successfully deleted NIC with ID %d from VM ID %d", portID, vmID)
}

// ImportState handles importing a resource.
func (r *VstackNicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Split the provided import ID into VM ID and Port ID
	ids := strings.Split(req.ID, "/")
	if len(ids) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in the format 'vm_id/port_id'.",
		)
		return
	}

	// Parse the VM ID and Port ID
	vmID, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid VM ID",
			fmt.Sprintf("Unable to parse VM ID '%s': %s", ids[0], err),
		)
		return
	}

	portID, err := strconv.ParseInt(ids[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Port ID",
			fmt.Sprintf("Unable to parse Port ID '%s': %s", ids[1], err),
		)
		return
	}

	// Set only `vm_id` and `id` during import

	resp.State.SetAttribute(ctx, path.Root("vm_id"), vmID)
	resp.State.SetAttribute(ctx, path.Root("id"), portID)
	//resp.Diagnostics.Append(resp.State.Set(ctx, map[string]any{
	//	"vm_id": types.Int64Value(vmID),
	//	"id":    types.Int64Value(portID),
	//})...)
}
