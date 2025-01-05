// Copyright (c) Ivan Brykalov, ivbrykalov@gmail.com
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"net/http"
	"terraform-provider-vstack/internal/helper"
	"terraform-provider-vstack/internal/models"
	"terraform-provider-vstack/internal/vstack_api"
)

// vStackVMProfileDataSource implements a data source for retrieving OS types and their profiles.
type VstackVMProfileDataSource struct {
	Client     *http.Client
	BaseURL    string
	AuthCookie string
}

// Ensure VstackVMProfileDataSource satisfies the Terraform interfaces.
var (
	_ datasource.DataSource              = &VstackVMProfileDataSource{}
	_ datasource.DataSourceWithConfigure = &VstackVMProfileDataSource{}
)

// NewVstackVMProfileDataSource initializes the data source.
func NewVstackVMProfileDataSource() datasource.DataSource {
	return &VstackVMProfileDataSource{}
}

// Metadata sets the name of the data source.
func (d *VstackVMProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_profile"
}

// Configure sets up the data source with provider-specific settings.
func (d *VstackVMProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*VStackProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *VStackProvider, got: %T", req.ProviderData),
		)
		return
	}

	d.Client = providerData.client
	d.BaseURL = providerData.Host
	d.AuthCookie = providerData.authCookie
}

// Schema defines the structure of the data source.
func (d *VstackVMProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"os_types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"profiles": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										Computed: true,
									},
									"name": schema.StringAttribute{
										Computed: true,
									},
									"description": schema.StringAttribute{
										Computed: true,
									},
									"min_size": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read sends a request to the "vm-profiles" method and converts the response into Terraform structures.
func (d *VstackVMProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// 1. Prepare the request payload.
	requestPayload := helper.BuildJSONRPCRequest("vm-profiles", nil)

	// 2. Send the request.
	apiResponse, err := vstack_api.VmProfiles(requestPayload, d.AuthCookie, d.BaseURL, d.Client)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving VM profiles", err.Error())
		return
	}

	// 3. Create the top-level model to store all OS types.
	var result models.VMProfileDataModel

	// 4. Convert the API response to Go structures.
	for _, osType := range apiResponse.Data {
		// Collect the list of profiles.
		var profiles []models.ProfileModel
		for _, p := range osType.Profiles {
			profiles = append(profiles, models.ProfileModel{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				MinSize:     p.MinSize,
			})
		}

		// Add to the result.
		result.OsTypes = append(result.OsTypes, models.OsTypeModel{
			ID:       osType.ID,
			Name:     osType.Name,
			Profiles: profiles,
		})
	}

	// 5. Save the result in the Terraform state.
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
