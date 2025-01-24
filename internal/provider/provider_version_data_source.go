package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &versionDataSource{}
	_ datasource.DataSourceWithConfigure = &versionDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewVersionDataSource() datasource.DataSource {
	return &versionDataSource{}
}

// coffeesDataSource is the data source implementation.
type versionDataSource struct {
	client *pve.PVE
}

// Metadata returns the data source type name.
func (d *versionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}
func (d *versionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Computed: true,
			},
			"repo_id": schema.StringAttribute{
				Computed: true,
			},
			"release": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *versionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state versionDataSourceModel

	version, err := d.client.GetVersion()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Version",
			err.Error(),
		)
		return
	}

	state.Release = types.StringValue(version.Release)
	state.Version = types.StringValue(version.Version)
	state.RepoID = types.StringValue(version.RepoID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *versionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pve.PVE)
	client.GetVersion()
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// versionDataSourceModel maps the data source schema data.
type versionDataSourceModel struct {
	Release types.String `tfsdk:"release"`
	Version types.String `tfsdk:"version"`
	RepoID  types.String `tfsdk:"repo_id"`
}

// versionModel maps version schema data.
type versionModel struct {
	Release types.String `tfsdk:"release"`
	Version types.String `tfsdk:"version"`
	RepoID  types.String `tfsdk:"repo_id"`
}
