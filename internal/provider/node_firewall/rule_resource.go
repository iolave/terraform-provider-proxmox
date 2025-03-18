package nodefirewall

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RuleResource{}
var _ resource.ResourceWithImportState = &RuleResource{}

func NewRuleResource() resource.Resource {
	return &RuleResource{}
}

// RuleResource defines the resource implementation.
type RuleResource struct {
	client *pve.PVE
}

// RuleResourceModel describes the resource data model.
type RuleResourceModel struct {
	ruleModel
	Node types.String `tfsdk:"node"`
}

func (r *RuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	name := "node_firewall_rule"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (r *RuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Node firewall rules resource",
		Description:         DESC_RULE,
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				Description: DESC_RULE_NODE,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: DESC_RULE_ID,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: DESC_RULE_ACTION,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_COMMENT,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_DEST,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dport": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_DPORT,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: DESC_RULE_ENABLE,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"icmp_type": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_ICMP,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"iface": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_IFACE,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_version": schema.Int64Attribute{
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"log": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_LOG,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"macro": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_MACRO,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// FIXME: POS NOT WORKING
			"pos": schema.Int64Attribute{
				Optional:    true,
				Description: DESC_RULE_POS,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"proto": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_PROTO,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_SOURCE,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sport": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_SPORT,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: DESC_RULE_TYPE,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *RuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	enable := 0
	if data.Enable.ValueBool() {
		enable = 1
	}

	apiReq := pve.CreateNodeFirewallRuleRequest{
		Action:      data.Action.ValueString(),
		Node:        data.Node.ValueString(),
		Type:        data.Type.ValueString(),
		Comment:     data.Comment.ValueString(),
		Destination: data.Destination.ValueString(),
		//Digest:          data.Digest.ValueString(),
		DestinationPort: data.DestinationPort.ValueString(),
		Enable:          enable,
		ICMPType:        data.ICMPType.ValueString(),
		Interface:       data.Interface.ValueString(),
		LogLevel:        pve.FirewallLogLevel(data.LogLevel.ValueString()),
		Macro:           data.Macro.ValueString(),
		// FIXME: POS NOT WORKING
		//Pos:             int(data.Pos.ValueInt64()),
		Proto:  data.Proto.ValueString(),
		Source: data.Source.ValueString(),
		Sport:  data.Sport.ValueString(),
	}

	id, err := r.client.Node.Firewall.NewRule(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rule, got error: %s", err))
		return
	}
	data.ID = types.StringValue(id)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	tflog.Debug(ctx, "GetRule args", map[string]any{
		"id":   data.ID.ValueString(),
		"node": data.Node.ValueString(),
	})
	remoteRule, err := r.client.Node.Firewall.GetRule(
		data.Node.ValueString(),
		data.ID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
		return
	}
	data.ID = types.StringValue(remoteRule.ID)
	data.Action = types.StringValue(remoteRule.Action)
	idx := strings.IndexRune(remoteRule.Comment, ']')
	comment := ""
	if len(remoteRule.Comment) >= idx+2 {
		comment = remoteRule.Comment[idx+2:]
	}
	data.Comment = types.StringValue(comment)

	data.Destination = types.StringValue(remoteRule.Destination)
	data.DestinationPort = types.StringValue(remoteRule.DestinationPort)
	enable := false
	if remoteRule.Enable == 1 {
		enable = true
	}
	data.Enable = types.BoolValue(enable)
	data.ICMPType = types.StringValue(remoteRule.ICMPType)
	data.Interface = types.StringValue(remoteRule.Interface)
	data.IPVersion = types.Int64Value(int64(remoteRule.IPVersion))
	data.LogLevel = types.StringValue(string(remoteRule.LogLevel))
	data.Macro = types.StringValue(remoteRule.Macro)
	// FIXME: POS NOT WORKING
	//data.Pos = types.Int64Value(int64(remoteRule.Pos))
	data.Proto = types.StringValue(remoteRule.Proto)
	data.Source = types.StringValue(remoteRule.Source)
	data.Sport = types.StringValue(remoteRule.Sport)
	data.Type = types.StringValue(remoteRule.Type)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RuleResourceModel
	var state RuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err := r.client.Node.Firewall.DeleteRule(data.Node.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
		return
	}

	enable := 0
	if data.Enable.ValueBool() {
		enable = 1
	}

	apiReq := pve.CreateNodeFirewallRuleRequest{
		Action:      data.Action.ValueString(),
		Node:        data.Node.ValueString(),
		Type:        data.Type.ValueString(),
		Comment:     data.Comment.ValueString(),
		Destination: data.Destination.ValueString(),
		//Digest:          data.Digest.ValueString(),
		DestinationPort: data.DestinationPort.ValueString(),
		Enable:          enable,
		ICMPType:        data.ICMPType.ValueString(),
		Interface:       data.Interface.ValueString(),
		LogLevel:        pve.FirewallLogLevel(data.LogLevel.ValueString()),
		Macro:           data.Macro.ValueString(),
		// FIXME: POS NOT WORKING
		// Pos:             int(data.Pos.ValueInt64()),
		Proto:  data.Proto.ValueString(),
		Source: data.Source.ValueString(),
		Sport:  data.Sport.ValueString(),
	}

	id, err := r.client.Node.Firewall.NewRule(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rule, got error: %s", err))
		return
	}
	data.ID = types.StringValue(id)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err := r.client.Node.Firewall.DeleteRule(data.Node.ValueString(), data.ID.ValueString()); err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "deleted a resource")
}

func (r *RuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
