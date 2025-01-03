package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-vstack/internal/helper"
)

// Ensure VStackProvider satisfies various provider interfaces.
var _ provider.Provider = &VStackProvider{}

// VStackProvider defines the provider implementation.
type VStackProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and run locally, and "test" when running acceptance
	// testing.
	version    string
	authCookie string
	client     *http.Client
	Host       string
}

// VStackProviderModel describes the provider data model.
type VStackProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata sets the type name and version for the provider.
func (p *VStackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vstack"
	resp.Version = p.version
}

// Schema defines the provider's configuration schema.
func (p *VStackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "API host for the Vstack provider.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for Vstack API.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for Vstack API.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure initializes the provider with the provided configuration.
// It authenticates with the Vstack API and stores the authentication cookie.
func (p *VStackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Starting provider configuration")
	fmt.Println("Starting provider configuration")

	var data VStackProviderModel

	// Retrieve configuration values from req.Config
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		fmt.Println("Error retrieving provider configuration")
		return
	}

	// Initialize HTTP client
	p.client = &http.Client{}
	p.authCookie = ""
	p.Host = data.Host.ValueString()

	// Prepare JSON-RPC request for authentication using helper function
	authReq := helper.BuildJSONRPCRequest("auth", map[string]interface{}{
		"username": data.Username.ValueString(),
		"password": data.Password.ValueString(),
	})

	body, err := json.Marshal(authReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling authentication request", err.Error())
		return
	}
	reqBody := bytes.NewBuffer(body)

	// Send authentication request
	authURL := fmt.Sprintf("%s/.api/V4/.req/", p.Host)
	authRequest, err := http.NewRequest("POST", authURL, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating authentication request", err.Error())
		return
	}
	authRequest.Header.Set("Content-Type", "application/json")

	// Execute authentication request
	authResponse, err := p.client.Do(authRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error authenticating with Vstack", err.Error())
		return
	}
	defer func() {
		if closeErr := authResponse.Body.Close(); closeErr != nil {
			tflog.Error(ctx, "Error closing authentication response body", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	// Decode authentication response
	var authResp struct {
		ID      string `json:"id"`
		JsonRPC string `json:"jsonrpc"`
		Result  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Cookie map[string]string `json:"cookie"`
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.NewDecoder(authResponse.Body).Decode(&authResp); err != nil {
		resp.Diagnostics.AddError("Error decoding authentication response", err.Error())
		return
	}

	// Check authentication result code
	if authResp.Result.Code != 1 {
		resp.Diagnostics.AddError("Authentication failed", authResp.Result.Message)
		return
	}

	// Store the received auth cookie
	p.authCookie = authResp.Result.Data.Cookie["APIEndpoint00"]

	// Retrieve cookie from response headers as a fallback
	for _, cookie := range authResponse.Cookies() {
		if cookie.Name == "APIEndpoint00" {
			p.authCookie = cookie.Value
			break
		}
	}
	if p.authCookie == "" {
		resp.Diagnostics.AddError("Authentication failed", "No auth cookie received")
		return
	}

	tflog.Info(ctx, "Provider configuration completed", map[string]any{"provider": p})
	// Pass the provider instance to be used in DataSources and Resources
	resp.DataSourceData = p
	resp.ResourceData = p
	fmt.Println("Provider successfully configured")
}

// Resources returns a list of resource constructors for the provider.
func (p *VStackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVstackVMResource,
		NewVstackNicResource,
	}
}

// DataSources returns a list of data source constructors for the provider.
func (p *VStackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource, // Uncomment if you have additional data sources
		NewVstackVMGetDataSource,
	}
}

// New initializes and returns a new instance of the VStackProvider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VStackProvider{
			version: version,
		}
	}
}
