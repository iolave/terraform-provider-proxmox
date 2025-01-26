package provider

import (
	"context"
	"os"
	"strconv"
	nodefirewall "terraform-provider-proxmox/internal/provider/node_firewall"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iolave/go-proxmox/pkg/cloudflare"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &proxmoxProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &proxmoxProvider{
			version: version,
		}
	}
}

// proxmoxProvider is the provider implementation.
type proxmoxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *proxmoxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox"
	resp.Version = p.version
}

// proxmoxProviderModel maps provider schema data to a Go type.
type proxmoxProviderModel struct {
	Host               types.String `tfsdk:"host"`
	Port               types.Int32  `tfsdk:"port"`
	User               types.String `tfsdk:"user"`
	TokenName          types.String `tfsdk:"token_name"`
	Token              types.String `tfsdk:"token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	CfClientId         types.String `tfsdk:"cf_client_id"`
	CfClientSecret     types.String `tfsdk:"cf_client_secret"`
}

// Schema defines the provider-level schema for configuration data.
func (p *proxmoxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"port": schema.Int32Attribute{
				Optional: true,
			},
			"user": schema.StringAttribute{
				Optional: true,
			},
			"token_name": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional: true,
			},
			"cf_client_id": schema.StringAttribute{
				Optional: true,
			},
			"cf_client_secret": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Configure prepares a Proxmox API client for data sources and resources.
func (p *proxmoxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config proxmoxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Proxmox API Host",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_HOST environment variable.",
		)
	}

	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown Proxmox API Port",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API port. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_PORT environment variable.",
		)
	}

	if config.User.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Unknown Proxmox API User",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API user. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_USER environment variable.",
		)
	}

	if config.TokenName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token_name"),
			"Unknown Proxmox API Token name",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API token name. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_TOKEN_NAME environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Proxmox API Token",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_TOKEN environment variable.",
		)
	}

	if config.InsecureSkipVerify.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("insecure_skip_verify"),
			"Unknown Proxmox API Insecure skip verify",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API Insecure Skip Verify option.",
		)
	}

	if config.CfClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("cf_client_id"),
			"Unknown Proxmox API cloudflare client id",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API Cloudflare client id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CF_CLIENT_ID environment variable.",
		)
	}

	if config.CfClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("cf_client_secret"),
			"Unknown Proxmox API cloudflare client secret",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API Cloudflare client secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CF_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("PROXMOX_HOST")
	var port int
	var err error
	if portEnv := os.Getenv("PROXMOX_PORT"); portEnv != "" {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("port"),
				"Invalid Proxmox API Port",
				"The provider cannot create the Proxmox API client as there is a non int value for the Proxmox API Port. "+
					"Set the port value in the configuration or use the PROXMOX_PORT environment variable. "+
					"If either is already set, ensure the value is a valid int.",
			)
		}
	}
	user := os.Getenv("PROXMOX_USER")
	tokenName := os.Getenv("PROXMOX_TOKEN_NAME")
	token := os.Getenv("PROXMOX_TOKEN")
	insecureSkipVerify := false
	cfClientId := os.Getenv("CF_CLIENT_ID")
	cfClientSecret := os.Getenv("CF_CLIENT_SECRET")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Port.IsNull() {
		port = int(config.Port.ValueInt32())
	}

	if !config.User.IsNull() {
		user = config.User.ValueString()
	}

	if !config.TokenName.IsNull() {
		tokenName = config.TokenName.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.CfClientId.IsNull() {
		cfClientId = config.CfClientId.ValueString()
	}

	if !config.CfClientSecret.IsNull() {
		cfClientSecret = config.CfClientSecret.ValueString()
	}

	if !config.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = config.InsecureSkipVerify.ValueBool()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Proxmox API Host",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API host. "+
				"Set the host value in the configuration or use the PROXMOX_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if port == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Missing Proxmox API Port",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API port. "+
				"Set the host value in the configuration or use the PROXMOX_PORT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if user == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Missing Proxmox API User",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API user. "+
				"Set the user value in the configuration or use the PROXMOX_USER environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if tokenName == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tokenName"),
			"Missing Proxmox API Token name",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API token name. "+
				"Set the token name value in the configuration or use the PROXMOX_TOKEN_NAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Proxmox API Token",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API token. "+
				"Set the token value in the configuration or use the PROXMOX_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if cfClientId == "" && cfClientSecret != "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cfClientId"),
			"Missing Proxmox API Cloudflare client id",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API client id. "+
				"Set the client id value in the configuration or use the CF_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if cfClientId != "" && cfClientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cfClientId"),
			"Missing Proxmox API Cloudflare client secret",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the Proxmox API client secret. "+
				"Set the client secret value in the configuration or use the CF_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Proxmox client using the configuration values
	pveConfig := pve.Config{
		Host:               host,
		Port:               port,
		InsecureSkipVerify: insecureSkipVerify,
	}
	if cfClientId != "" && cfClientSecret != "" {
		cfServiceToken := cloudflare.NewServiceToken(cfClientId, cfClientSecret)
		pveConfig.CfServiceToken = cfServiceToken
	}

	creds := pve.NewTokenCreds(user, tokenName, token)
	pve, err := pve.NewWithCredentials(pveConfig, creds)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Proxmox API Client",
			"An unexpected error occurred when creating the Proxmox API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}

	// Make the Proxmox client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = pve
	resp.ResourceData = pve
}

// DataSources defines the data sources implemented in the provider.
func (p *proxmoxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVersionDataSource,
		nodefirewall.NewRulesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *proxmoxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		nodefirewall.NewRulesResource,
	}
}
