package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
	"terraform-provider-vstack/internal/helper"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// Ensure VstackVMGetDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &VstackVMGetDataSource{}
var _ datasource.DataSourceWithConfigure = &VstackVMGetDataSource{}

// VstackVMGetDataSource defines the implementation of the VM Get data source.
type VstackVMGetDataSource struct {
	Client     *http.Client
	BaseURL    string
	AuthCookie string
}

// NewVstackVMGetDataSource initializes the VM Get data source.
func NewVstackVMGetDataSource() datasource.DataSource {
	return &VstackVMGetDataSource{}
}

// Metadata sets the type name for the data source.
func (d *VstackVMGetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_get"
}

// Configure configures the data source with the provider's settings.
func (d *VstackVMGetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Info(ctx, "Starting data source configuration")

	// Check if provider data is available
	if req.ProviderData == nil {
		tflog.Warn(ctx, "Provider data is not configured")
		return
	}

	// Assert that provider data is of type *VStackProvider
	if providerData, ok := req.ProviderData.(*VStackProvider); ok {
		d.Client = providerData.client
		d.BaseURL = providerData.Host
		d.AuthCookie = providerData.authCookie
		tflog.Info(ctx, "Data source configured successfully", map[string]any{
			"BaseURL":    d.BaseURL,
			"AuthCookie": d.AuthCookie,
		})
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provider.VStackProvider, got %T", req.ProviderData),
		)
	}
}
func (d *VstackVMGetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the virtual machine.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the virtual machine.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the virtual machine.",
				Optional:    true,
				Computed:    true,
			},
			"cpus": schema.Int64Attribute{
				Description: "Number of CPUs assigned to the virtual machine.",
				Computed:    true,
			},
			"ram": schema.Int64Attribute{
				Description: "Amount of RAM in Mega bytes for the virtual machine.",
				Computed:    true,
			},
			"cpu_priority": schema.Int64Attribute{
				Description: "CPU priority of the virtual machine (1-20).",
				Computed:    true,
			},
			"boot_media": schema.Int64Attribute{
				Description: "ID of the boot media.",
				Computed:    true,
			},
			"vcpu_class": schema.Int64Attribute{
				Description: "Class of the vCPU for the virtual machine.",
				Computed:    true,
			},
			"os_type": schema.Int64Attribute{
				Description: "Operating system type for the virtual machine.",
				Computed:    true,
			},
			"os_profile": schema.StringAttribute{
				Description: "Operating system profile for the virtual machine.",
				Computed:    true,
			},
			"vdc_id": schema.Int64Attribute{
				Description: "Virtual Data Center ID for the virtual machine.",
				Computed:    true,
			},
			"node": schema.Int64Attribute{
				Description: "Node on which the VM is running.",
				Computed:    true,
			},
			"uefi": schema.StringAttribute{
				Description: "UEFI firmware path.",
				Computed:    true,
			},
			"create_completed": schema.Int64Attribute{
				Description: "Indicates if the VM creation is completed.",
				Computed:    true,
			},
			"locked": schema.Int64Attribute{
				Description: "Indicates if the VM is locked.",
				Computed:    true,
			},
			"root_dataset": schema.StringAttribute{
				Description: "Root dataset ID of the VM.",
				Computed:    true,
			},
			"root_dataset_name": schema.StringAttribute{
				Description: "Root dataset name of the VM.",
				Computed:    true,
			},
			"pool_selector": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The pool where the virtual machine resides.",
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
			},
			"guest": schema.SingleNestedAttribute{
				Description: "Guest customization for the VM.",
				Optional:    true,
				Computed:    true,

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
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"ssh_authorized_keys": schema.ListAttribute{
									Description: "SSH public keys.",
									ElementType: types.StringType,
									Optional:    true,
									Computed:    true,
								},
								"password": schema.StringAttribute{
									Description: "Password for the user.",
									Optional:    true,
									Computed:    true,
								},
							},
						},
					},
					"ssh_password_auth": schema.Int64Attribute{
						Description: "Enables or disables SSH password authentication.",
						Computed:    true,
					},
					"resolver": schema.SingleNestedAttribute{
						Description: "DNS resolver settings for the guest OS.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"name_server": schema.ListAttribute{
								Description: "DNS name servers.",
								ElementType: types.StringType,
								Computed:    true,
							},
							"search": schema.StringAttribute{
								Description: "DNS search domain.",
								Computed:    true,
							},
						},
					},
					"boot_cmds": schema.ListAttribute{
						Description: "List of boot commands for the guest OS.",
						ElementType: types.StringType,
						Computed:    true,
					},
					"run_cmds": schema.ListAttribute{
						Description: "List of commands to run in the guest OS.",
						ElementType: types.StringType,
						Computed:    true,
					},
					"hostname": schema.StringAttribute{
						Description: "Hostname for the guest OS.",
						Computed:    true,
					},
				},
			},
			"disks": schema.ListNestedAttribute{
				Description: "List of disks attached to the virtual machine.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"guid": schema.StringAttribute{
							Description: "UUID Disk.",
							Computed:    true,
						},
						"size": schema.Int64Attribute{
							Description: "Size of the disk in Gigabytes.",
							Computed:    true,
						},
						"slot": schema.Int64Attribute{
							Description: "Slot number for the disk.",
							Computed:    true,
						},
						"iops_limit": schema.Int64Attribute{
							Description: "IOPS limit for the disk.",
							Computed:    true,
						},
						"mbps_limit": schema.Int64Attribute{
							Description: "Mbps limit for the disk.",
							Computed:    true,
						},
						"label": schema.StringAttribute{
							Description: "Label for the disk.",
							Computed:    true,
						},
						"sector_size": schema.SingleNestedAttribute{
							Description: "Sector size for the disk.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"logical": schema.Int64Attribute{
									Description: "Logical sector size.",
									Computed:    true,
								},
								"physical": schema.Int64Attribute{
									Description: "Physical sector size.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read retrieves data for the VM Get data source.
func (d *VstackVMGetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state models.VMResourceModel

	// Get the VM ID from the configuration
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the JSON-RPC request payload using the helper function
	requestPayload := helper.BuildJSONRPCRequest("vm-get", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	// Call vstack-api and get VM info
	apiResponse, err := vstack_api.VmGet(requestPayload, d.AuthCookie, d.BaseURL, d.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error on Read VM Data in vstack_api.VmGet", err.Error())
		return
	}

	// Map response fields to the state using helper function
	updatedState, err := helper.MapRespToState(apiResponse, state)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping response to state", err.Error())
		return
	}

	// Set the updated state
	diags = resp.State.Set(ctx, updatedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
