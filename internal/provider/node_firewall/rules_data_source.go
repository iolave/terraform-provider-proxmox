package nodefirewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &rulesDataSource{}
	_ datasource.DataSourceWithConfigure = &rulesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewRulesDataSource() datasource.DataSource {
	return &rulesDataSource{}
}

// coffeesDataSource is the data source implementation.
type rulesDataSource struct {
	client *pve.PVE
}

// Metadata returns the data source type name.
func (d *rulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	name := "node_firewall_rules"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (d *rulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ruleSchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"action":      schema.StringAttribute{Computed: true},
			"comment":     schema.StringAttribute{Computed: true},
			"destination": schema.StringAttribute{Computed: true},
			"dport":       schema.Int64Attribute{Computed: true},
			"enable":      schema.BoolAttribute{Computed: true},
			"icmp_type":   schema.StringAttribute{Computed: true},
			"iface":       schema.StringAttribute{Computed: true},
			"ip_version":  schema.Int64Attribute{Computed: true},
			"log":         schema.StringAttribute{Computed: true},
			"macro":       schema.StringAttribute{Computed: true},
			"pos":         schema.Int64Attribute{Computed: true},
			"proto":       schema.StringAttribute{Computed: true},
			"source":      schema.StringAttribute{Computed: true},
			"sport":       schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				Required: true,
			},
			"rules": schema.ListNestedAttribute{
				NestedObject: ruleSchema,
				Computed:     true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *rulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rulesDataSourceModel
	resp.State.Get(ctx, &state)
	state.Rules = []ruleModel{}
	tflog.Info(ctx, "reading node firewall rules", map[string]interface{}{"node": state.Node.ValueString()})

	rules, err := d.client.Node.Firewall.GetRules(state.Node.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox Rules",
			err.Error(),
		)
		return
	}

	for _, r := range rules {
		rule := ruleModel{}

		rule.ID = types.StringValue(r.ID)
		rule.Action = types.StringValue(r.Action)
		rule.Comment = types.StringValue(r.Comment)
		rule.Destination = types.StringValue(r.Destination)
		rule.DestinationPort = types.Int64Value(int64(r.DestinationPort))
		switch r.Enable {
		case 0:
			rule.Enable = types.BoolValue(false)
		case 1:
			rule.Enable = types.BoolValue(true)
		default:
			resp.Diagnostics.AddError(
				"Unable to parse proxmox node firewall rule",
				fmt.Sprintf("Expected Enabled property value to be 1 or 0, but got %d", r.Enable),
			)
			return
		}
		rule.ICMPType = types.StringValue(r.ICMPType)
		rule.Interface = types.StringValue(r.Interface)
		rule.IPVersion = types.Int64Value(int64(r.IPVersion))
		rule.LogLevel = types.StringValue(string(r.LogLevel))
		rule.Macro = types.StringValue(r.Macro)
		rule.Pos = types.Int64Value(int64(r.Pos))
		rule.Proto = types.StringValue(r.Proto)
		rule.Source = types.StringValue(r.Source)
		rule.Sport = types.StringValue(r.Sport)
		rule.Type = types.StringValue(r.Type)

		state.Rules = append(state.Rules, rule)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *rulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pve.PVE)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *pve.PVE, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// rulesDataSourceModel maps the data source schema data.
type rulesDataSourceModel struct {
	Node  types.String `tfsdk:"node"`
	Rules []ruleModel  `tfsdk:"rules"`
}

// rulesModel maps rule schema data.
type ruleModel struct {
	ID              types.String `tfsdk:"id"`
	Action          types.String `tfsdk:"action"`
	Comment         types.String `tfsdk:"comment"`
	Destination     types.String `tfsdk:"destination"`
	DestinationPort types.Int64  `tfsdk:"dport"`
	Enable          types.Bool   `tfsdk:"enable"`
	ICMPType        types.String `tfsdk:"icmp_type"`
	Interface       types.String `tfsdk:"iface"`
	IPVersion       types.Int64  `tfsdk:"ip_version"`
	LogLevel        types.String `tfsdk:"log"`
	Macro           types.String `tfsdk:"macro"`
	Pos             types.Int64  `tfsdk:"pos"`
	Proto           types.String `tfsdk:"proto"`
	Source          types.String `tfsdk:"source"`
	Sport           types.String `tfsdk:"sport"`
	Type            types.String `tfsdk:"type"`
}
