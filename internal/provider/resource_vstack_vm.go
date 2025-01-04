// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"log"
	"net/http"
	"strconv"
	"strings"
	"terraform-provider-vstack/internal/helper"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

type VstackVMResource struct {
	Client     *http.Client
	BaseURL    string
	AuthCookie string
}

func NewVstackVMResource() resource.Resource {
	return &VstackVMResource{}
}

func (r *VstackVMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VstackVMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if providerData, ok := req.ProviderData.(*VStackProvider); ok {
		r.Client = providerData.client
		r.AuthCookie = providerData.authCookie
		r.BaseURL = providerData.Host
		if r.Client == nil || r.BaseURL == "" {
			resp.Diagnostics.AddError(
				"Client Initialization Error",
				"The HTTP client or baseURL was not properly initialized.",
			)
		}
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			"The provider data was not of the expected *VStackProvider type.",
		)
	}
}

func (r *VstackVMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the virtual machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the virtual machine.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the virtual machine.",
				Optional:    true,
			},
			"cpus": schema.Int64Attribute{
				Description: "Number of CPUs assigned to the virtual machine.",
				Required:    true,
			},
			"ram": schema.Int64Attribute{
				Description: "Amount of RAM in Mega bytes for the virtual machine.",
				Required:    true,
			},
			"cpu_priority": schema.Int64Attribute{
				Description: "CPU priority of the virtual machine (1-20).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"boot_media": schema.Int64Attribute{
				Description: "ID of the boot media.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"vcpu_class": schema.Int64Attribute{
				Description: "Class of the vCPU for the virtual machine.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"os_type": schema.Int64Attribute{
				Description: "Operating system type for the virtual machine.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"os_profile": schema.StringAttribute{
				Description: "Operating system profile for the virtual machine.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vdc_id": schema.Int64Attribute{
				Description: "Virtual Data Center ID for the virtual machine.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"node": schema.Int64Attribute{
				Description: "Node on which the VM is running.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"uefi": schema.StringAttribute{
				Description: "UEFI firmware path.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_completed": schema.Int64Attribute{
				Description: "Indicates if the VM creation is completed.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"locked": schema.Int64Attribute{
				Description: "Indicates if the VM is locked.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"root_dataset": schema.StringAttribute{
				Description: "Root dataset ID of the VM.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"root_dataset_name": schema.StringAttribute{
				Description: "Root dataset name of the VM.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool_selector": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The pool where the virtual machine resides.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.Int64Attribute{
				Description: "Indicates the status of the VM.",
				Computed:    true,
			},
			"admin_status": schema.Int64Attribute{
				Description: "Administrative status of the VM.",
				Computed:    true,
			},
			"oper_status": schema.Int64Attribute{
				Description: "Operational status of the VM.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "Action to perform on the VM (e.g., 'start', 'stop').",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"guest": schema.SingleNestedAttribute{
				Description: "Guest customization for the VM.",
				Required:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"ram_used": schema.Int64Attribute{
						Description: "RAM used by the guest operating system in MB.",
						Computed:    true,
					},
					"ram_balloon_performed": schema.Int64Attribute{
						Description: "RAM used by the guest operating system in MB.",
						Computed:    true,
					},
					"ram_balloon_requested": schema.Int64Attribute{
						Description: "RAM used by the guest operating system in MB.",
						Computed:    true,
					},
					"users": schema.MapNestedAttribute{
						Description: "List of users in the guest OS.",
						Required:    true,
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.RequiresReplace(),
							mapplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"ssh_authorized_keys": schema.ListAttribute{
									Description: "SSH public keys.",
									ElementType: types.StringType,
									Optional:    true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.RequiresReplace(),
									},
								},
								"password": schema.StringAttribute{
									Description: "Password for the user.",
									Required:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.RequiresReplace(),
									},
								},
							},
						},
					},
					"ssh_password_auth": schema.Int64Attribute{
						Description: "Enables or disables SSH password authentication.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
							int64planmodifier.RequiresReplace(),
						},
					},
					"resolver": schema.SingleNestedAttribute{
						Description: "DNS resolver settings for the guest OS.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"name_server": schema.ListAttribute{
								Description: "DNS name servers.",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
									listplanmodifier.RequiresReplace(),
								},
							},
							"search": schema.StringAttribute{
								Description: "DNS search domain.",
								Optional:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"boot_cmds": schema.ListAttribute{
						Description: "List of boot commands for the guest OS.",
						ElementType: types.StringType,
						Optional:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
							listplanmodifier.RequiresReplace(),
						},
					},
					"run_cmds": schema.ListAttribute{
						Description: "List of commands to run in the guest OS.",
						ElementType: types.StringType,
						Optional:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
							listplanmodifier.RequiresReplace(),
						},
					},
					"hostname": schema.StringAttribute{
						Description: "Hostname for the guest OS.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"disks": schema.ListNestedAttribute{
				Description: "List of disks attached to the virtual machine.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"guid": schema.StringAttribute{
							Description: "UUID Disk.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"size": schema.Int64Attribute{
							Description: "Size of the disk in Gigabytes.",
							Required:    true,
						},
						"slot": schema.Int64Attribute{
							Description: "Slot number for the disk.",
							Required:    true,
						},
						"iops_limit": schema.Int64Attribute{
							Description: "IOPS limit for the disk.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"mbps_limit": schema.Int64Attribute{
							Description: "Mbps limit for the disk.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"label": schema.StringAttribute{
							Description: "Label for the disk.",
							Optional:    true,
							Computed:    true,
						},
						"sector_size": schema.SingleNestedAttribute{
							Description: "Sector size for the disk.",
							Optional:    true,
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"logical": schema.Int64Attribute{
									Description: "Logical sector size.",
									Optional:    true,
									Computed:    true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"physical": schema.Int64Attribute{
									Description: "Physical sector size.",
									Optional:    true,
									Computed:    true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
							},
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

// Create: creates a VM and manages its state (start/stop) as needed.
func (r *VstackVMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 1. Retrieve the plan from the request
	var plan models.VMResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Apply default sector size for disks
	for i := range plan.Disks {
		helper.ApplyDefaultSectorSize(&plan.Disks[i])
	}

	// 3. Prepare the guest payload for VM creation
	guestPayload := make(map[string]interface{})

	// Only access plan.Guest if it is non-nil

	// Hostname
	hostname := plan.Guest.Hostname.ValueString()
	if hostname != "" {
		guestPayload["hostname"] = hostname
	}

	// BootCmds
	if !plan.Guest.BootCmds.IsNull() && !plan.Guest.BootCmds.IsUnknown() {
		bootCmds := make([]string, 0)
		var bootCmdsValues []attr.Value
		if err := plan.Guest.BootCmds.ElementsAs(ctx, &bootCmdsValues, false); err == nil {
			for _, cmd := range bootCmdsValues {
				stringVal, ok := cmd.(basetypes.StringValue)
				if !ok {
					log.Printf("Warning: BootCmd is not basetypes.StringValue, skipping.")
					continue
				}
				val := stringVal.ValueString()
				if val != "" {
					bootCmds = append(bootCmds, val)
				}
			}
		} else {
			log.Printf("Error retrieving BootCmds elements: %v", err)
		}
		if len(bootCmds) > 0 {
			guestPayload["boot_cmds"] = bootCmds
		}
	}

	// RunCmds
	if !plan.Guest.RunCmds.IsNull() && !plan.Guest.RunCmds.IsUnknown() {
		runCmds := make([]string, 0)
		var runCmdsValues []attr.Value
		if err := plan.Guest.RunCmds.ElementsAs(ctx, &runCmdsValues, false); err == nil {
			for _, cmd := range runCmdsValues {
				stringVal, ok := cmd.(basetypes.StringValue)
				if !ok {
					log.Printf("Warning: RunCmd is not basetypes.StringValue, skipping.")
					continue
				}
				val := stringVal.ValueString()
				if val != "" {
					runCmds = append(runCmds, val)
				}
			}
		} else {
			log.Printf("Error retrieving RunCmds elements: %v", err)
		}
		if len(runCmds) > 0 {
			guestPayload["run_cmds"] = runCmds
		}
	}

	// SSH password authentication
	sshPasswordAuth := plan.Guest.SSHPasswordAuth.ValueInt64()
	if sshPasswordAuth != 0 {
		guestPayload["ssh_password_auth"] = sshPasswordAuth
	}

	// Resolver
	if plan.Guest.Resolver != nil {
		resolverPayload := make(map[string]interface{})

		// Name servers
		if len(plan.Guest.Resolver.NameServers) > 0 {
			nsList := make([]string, 0, len(plan.Guest.Resolver.NameServers))
			for _, ns := range plan.Guest.Resolver.NameServers {
				val := ns.ValueString()
				if val != "" {
					nsList = append(nsList, val)
				}
			}
			if len(nsList) > 0 {
				resolverPayload["name_server"] = nsList
			}
		}

		// Search domain
		searchDomain := plan.Guest.Resolver.Search.ValueString()
		if searchDomain != "" {
			resolverPayload["search"] = searchDomain
		}

		// Add resolver payload only if not empty
		if len(resolverPayload) > 0 {
			guestPayload["resolver"] = resolverPayload
		}
	}

	// Users

	usersPayload := make(map[string]interface{})
	for username, user := range plan.Guest.Users {
		userPayload := make(map[string]interface{})

		// SSH authorized keys
		if len(user.SSHPublicKeys) > 0 {
			sshKeys := make([]string, 0, len(user.SSHPublicKeys))
			for _, key := range user.SSHPublicKeys {
				keyVal := key.ValueString()
				if keyVal != "" {
					sshKeys = append(sshKeys, keyVal)
				}
			}
			if len(sshKeys) > 0 {
				userPayload["ssh-authorized-keys"] = sshKeys
			}
		}

		// Password
		pwVal := user.Password.ValueString()
		if pwVal != "" {
			userPayload["password"] = pwVal
		}

		// Add user only if fields are non-empty
		if len(userPayload) > 0 {
			usersPayload[username] = userPayload
		}
	}
	if len(usersPayload) > 0 {
		guestPayload["users"] = usersPayload
	}

	// 4. Prepare the request for VM creation
	params := map[string]interface{}{
		"name":          plan.Name.ValueString(),
		"cpus":          plan.CPUs.ValueInt64(),
		"ram":           helper.ConvertMbToBytes(plan.RAM.ValueInt64()),
		"boot_media":    plan.BootMedia.ValueInt64(),
		"vcpu_class":    plan.VcpuClass.ValueInt64(),
		"os_type":       plan.OsType.ValueInt64(),
		"os_profile":    plan.OsProfile.ValueString(),
		"vdc_id":        plan.VdcID.ValueInt64(),
		"pool_selector": plan.PoolSelector.ValueString(),
		"disks":         helper.FormatDisks(plan.Disks),
		"description":   plan.Description.ValueString(),
	}

	// Attach guest payload only if we have data
	if len(guestPayload) > 0 {
		params["guest"] = guestPayload
	}

	if plan.CPUPriority.IsNull() || plan.CPUPriority.IsUnknown() {
		params["cpu_priority"] = 1
	} else {
		params["cpu_priority"] = plan.CPUPriority.ValueInt64()
	}

	requestCreatePayload := helper.BuildJSONRPCRequest("vms-create", params)

	// 5. Call the API to create the VM
	apiCreateResponse, err := vstack_api.VmCreate(requestCreatePayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error on vstack_api.VmCreate func", err.Error())
		return
	}

	// 6. Retrieve the ID of the created VM
	vmID := apiCreateResponse.Data.ID
	if vmID == 0 {
		resp.Diagnostics.AddError("Invalid VM ID", "VM ID is null or zero after creation.")
		return
	}

	// 7. Retrieve the mutex for the new VM and lock it
	mu, lockErr := helper.GetVMLock(vmID)
	if lockErr != nil {
		resp.Diagnostics.AddError("Error on GetVMLock", lockErr.Error())
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// 8. Manage VM state (start/stop) based on plan.Action
	action := strings.ToLower(plan.Action.ValueString())
	switch action {
	case "start":
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
			resp.Diagnostics.AddError("Error starting VM", err.Error())
			return
		}
	case "stop":
		// Check if we need to start before we can stop
		isRunning, err := helper.CheckIfVMIsRunning(r.Client, r.AuthCookie, r.BaseURL, vmID)
		if err != nil {
			resp.Diagnostics.AddError("Error checking VM status", err.Error())
			return
		}

		// If newly created or not running, start + stop
		if apiCreateResponse.Data.OperStatus == helper.Status.Created || !isRunning {
			if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
				resp.Diagnostics.AddError("Error starting VM before stopping", err.Error())
				return
			}
		}
		// Stop
		if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "stop"); err != nil {
			resp.Diagnostics.AddError("Error stopping VM", err.Error())
			return
		}
	case "":
		// No action specified, do nothing
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unsupported action '%s'. Supported actions are 'start' or 'stop'.", action))
		return
	}

	// 9. Retrieve the full VM details
	requestReadVMPayload := helper.BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": vmID,
	})

	apiResponse, err := vstack_api.VmGet(requestReadVMPayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error on vstack_api.VmGet func", err.Error())
		return
	}

	// 10. Map the API response to Terraform state
	updatedState, mapErr := helper.MapRespToState(apiResponse, plan)
	if mapErr != nil {
		resp.Diagnostics.AddError("Error mapping response to state in Create", mapErr.Error())
		return
	}
	plan = updatedState

	// 11. Save the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 12. Log successful VM creation
	log.Printf("Successfully created VM with ID %d", vmID)
}

// Read: retrieves the current state of the VM from the API and updates the Terraform state.
func (r *VstackVMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.VMResourceModel

	// Retrieve the current state for the VM ID
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueInt64()
	if vmID == 0 {
		resp.Diagnostics.AddError("Invalid VM ID", "VM ID is null or zero.")
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

	// Call the vm-get API to get the latest VM information
	requestPayload := helper.BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": vmID,
	})

	apiResponse, err := vstack_api.VmGet(requestPayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error on vstack_api.VmGet func", err.Error())
		return
	}

	// Map the API response to Terraform state
	updatedState, mapErr := helper.MapRespToState(apiResponse, state)
	if mapErr != nil {
		resp.Diagnostics.AddError("Error mapping response to state in Read", mapErr.Error())
		return
	}
	state = updatedState

	// Set the updated state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log successful VM state read
	log.Printf("Successfully read VM state for VM ID %d", vmID)
}

// Update: updates the VM parameters and manages its state (start/stop).
func (r *VstackVMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.VMResourceModel

	// Retrieve the plan and current state
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

	vmID := plan.ID.ValueInt64()
	if vmID == 0 {
		resp.Diagnostics.AddError("Invalid VM ID", "VM ID is null or zero.")
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

	// 1. Handle disk updates
	diskDiags := r.UpdateDisks(ctx, &plan, &state)
	resp.Diagnostics.Append(diskDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Collect changed VM parameters
	vmParams := make(map[string]interface{})

	if plan.Name.ValueString() != state.Name.ValueString() {
		vmParams["name"] = plan.Name.ValueString()
	}
	if plan.Description.ValueString() != state.Description.ValueString() {
		vmParams["description"] = plan.Description.ValueString()
	}
	if plan.CPUs.ValueInt64() != state.CPUs.ValueInt64() {
		vmParams["cpus"] = plan.CPUs.ValueInt64()
	}
	if plan.RAM.ValueInt64() != state.RAM.ValueInt64() {
		vmParams["ram"] = helper.ConvertMbToBytes(plan.RAM.ValueInt64())
	}
	if plan.CPUPriority.ValueInt64() != state.CPUPriority.ValueInt64() {
		vmParams["cpu_priority"] = plan.CPUPriority.ValueInt64()
	}
	if plan.BootMedia.ValueInt64() != state.BootMedia.ValueInt64() {
		vmParams["boot_media"] = plan.BootMedia.ValueInt64()
	}
	if plan.VcpuClass.ValueInt64() != state.VcpuClass.ValueInt64() {
		vmParams["vcpu_class"] = plan.VcpuClass.ValueInt64()
	}
	if plan.OsType.ValueInt64() != state.OsType.ValueInt64() {
		vmParams["os_type"] = plan.OsType.ValueInt64()
	}
	if plan.OsProfile.ValueString() != state.OsProfile.ValueString() {
		vmParams["os_profile"] = plan.OsProfile.ValueString()
	}
	if plan.VdcID.ValueInt64() != state.VdcID.ValueInt64() {
		vmParams["vdc_id"] = plan.VdcID.ValueInt64()
	}
	if plan.PoolSelector.ValueString() != state.PoolSelector.ValueString() {
		vmParams["pool_selector"] = plan.PoolSelector.ValueString()
	}

	// 3. Update VM parameters if there are changes
	if len(vmParams) > 0 {
		requestPayload := helper.BuildJSONRPCRequest("vm-set", map[string]interface{}{
			"id":        vmID,
			"vm_params": vmParams,
		})
		if _, err := vstack_api.VmSet(requestPayload, r.AuthCookie, r.BaseURL, r.Client); err != nil {
			resp.Diagnostics.AddError("Error updating VM parameters", err.Error())
			return
		}
	}

	// 4. Manage VM state (start/stop) based on the plan
	action := strings.ToLower(plan.Action.ValueString())
	if action != "" {
		switch action {
		case "start":
			// Execute the "start" action
			if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
				resp.Diagnostics.AddError("Error starting VM", err.Error())
				return
			}
		case "stop":
			// Check the current status of the VM
			isRunning, err := helper.CheckIfVMIsRunning(r.Client, r.AuthCookie, r.BaseURL, vmID)
			if err != nil {
				resp.Diagnostics.AddError("Error checking VM status", err.Error())
				return
			}

			if state.OperStatus.ValueInt64() == helper.Status.Created || !isRunning {
				// If status is "Created" or VM is not running, start and then stop
				if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "start"); err != nil {
					resp.Diagnostics.AddError("Error starting VM before stopping", err.Error())
					return
				}
			}

			// Stop the VM
			if err := helper.PerformAction(r.Client, r.AuthCookie, r.BaseURL, vmID, "stop"); err != nil {
				resp.Diagnostics.AddError("Error stopping VM", err.Error())
				return
			}
		default:
			resp.Diagnostics.AddError("Invalid Action", fmt.Sprintf("Unsupported action '%s'. Supported actions are 'start' and 'stop'.", action))
			return
		}

		// Reset the action in the plan to avoid saving it
		plan.Action = types.StringValue("")
	}

	// 5. Retrieve the full information of the VM to set the state
	requestReadVMPayload := helper.BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": vmID,
	})

	apiResponse, err := vstack_api.VmGet(requestReadVMPayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error on vstack_api.VmGet func", err.Error())
		return
	}

	// 6. Map the API response to Terraform state
	updatedState, mapErr := helper.MapRespToState(apiResponse, state)
	if mapErr != nil {
		resp.Diagnostics.AddError("Error mapping response to state in Update", mapErr.Error())
		return
	}
	state = updatedState

	// 7. Set the updated state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log successful VM update
	log.Printf("Successfully updated VM with ID %d", vmID)
}

// Delete: deletes the VM.
func (r *VstackVMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.VMResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueInt64()
	if vmID == 0 {
		resp.Diagnostics.AddError("Invalid VM ID", "VM ID is null or zero.")
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
			resp.Diagnostics.AddError("Error stopping VM", err.Error())
			return
		}
	}

	// 3. Call the API to delete the VM
	requestPayload := helper.BuildJSONRPCRequest("vms-remove", map[string]interface{}{
		"id":     vmID,
		"vdc_id": state.VdcID.ValueInt64(),
	})
	_, err = vstack_api.VmRemove(requestPayload, r.AuthCookie, r.BaseURL, r.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VM", err.Error())
		return
	}

	// 4. Remove the resource from Terraform state
	resp.State.RemoveResource(ctx)

	// Log successful VM deletion
	log.Printf("Successfully deleted VM with ID %d", vmID)
}

// ImportState handles the import logic for the VM resource.
func (r *VstackVMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID provided in the terraform import command
	importID := req.ID

	// Validate the import ID if necessary (e.g., ensure it's an integer)
	vmID, err := strconv.ParseInt(importID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected an integer VM ID, got: %s", importID),
		)
		return
	}

	// Set the ID attribute to the imported VM ID
	resp.State.SetAttribute(ctx, path.Root("id"), vmID)

	// Terraform will automatically call the Read method to populate other attributes
}
