package nodefirewall

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RulesResource{}
var _ resource.ResourceWithImportState = &RulesResource{}

func NewRulesResource() resource.Resource {
	return &RulesResource{}
}

// RulesResource defines the resource implementation.
type RulesResource struct {
	client *pve.PVE
}

// RulesResourceModel describes the resource data model.
type RulesResourceModel struct {
	Node  types.String `tfsdk:"node"`
	Rules []ruleModel  `tfsdk:"rules"`
}

func (r *RulesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	name := "node_firewall_rules"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (r *RulesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	ruleSchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: DESC_RULE_ID,
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: DESC_RULE_ACTION,
			},
			"comment": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_COMMENT,
			},
			"destination": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_DEST,
			},
			"dport": schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(0),
				Description: DESC_RULE_DPORT,
			},
			"enable": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: DESC_RULE_ENABLE,
			},
			"icmp_type": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_ICMP,
			},
			"iface": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_IFACE,
			},
			"ip_version": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				Default:  int64default.StaticInt64(4),
			},
			"log": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_LOG,
			},
			"macro": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_MACRO,
			},
			// FIXME: POS NOT WORKING
			"pos": schema.Int64Attribute{
				Optional:    true,
				Description: DESC_RULE_POS,
			},
			"proto": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_PROTO,
			},
			"source": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_SOURCE,
			},
			"sport": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: DESC_RULE_SPORT,
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: DESC_RULE_TYPE,
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Node firewall rules resource",
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				Required:    true,
				Description: DESC_RULE_NODE,
			},
			"rules": schema.ListNestedAttribute{
				NestedObject: ruleSchema,
				Required:     true,
				Description:  DESC_RULES,
			},
		},
	}
}

func (r *RulesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	for i, rule := range data.Rules {
		enable := 0
		if rule.Enable.ValueBool() {
			enable = 1
		}

		apiReq := pve.CreateNodeFirewallRuleRequest{
			Action:      rule.Action.ValueString(),
			Node:        data.Node.ValueString(),
			Type:        rule.Type.ValueString(),
			Comment:     rule.Comment.ValueString(),
			Destination: rule.Destination.ValueString(),
			//Digest:          rule.Digest.ValueString(),
			DestinationPort: int(rule.DestinationPort.ValueInt64()),
			Enable:          enable,
			ICMPType:        rule.ICMPType.ValueString(),
			Interface:       rule.Interface.ValueString(),
			LogLevel:        pve.FirewallLogLevel(rule.LogLevel.ValueString()),
			Macro:           rule.Macro.ValueString(),
			// FIXME: POS NOT WORKING
			//Pos:             int(rule.Pos.ValueInt64()),
			Proto:  rule.Proto.ValueString(),
			Source: rule.Source.ValueString(),
			Sport:  rule.Sport.ValueString(),
		}

		id, err := r.client.Node.Firewall.NewRule(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
			return
		}
		data.Rules[i].ID = types.StringValue(id)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RulesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	for i, rule := range data.Rules {
		remoteRule, err := r.client.Node.Firewall.GetRule(
			data.Node.ValueString(),
			rule.ID.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
			return
		}
		data.Rules[i].ID = types.StringValue(remoteRule.ID)
		data.Rules[i].Action = types.StringValue(remoteRule.Action)
		idx := strings.IndexRune(remoteRule.Comment, ']')
		comment := remoteRule.Comment[idx+2:]
		data.Rules[i].Comment = types.StringValue(comment)

		data.Rules[i].Destination = types.StringValue(remoteRule.Destination)
		data.Rules[i].DestinationPort = types.Int64Value(int64(remoteRule.DestinationPort))
		enable := false
		if remoteRule.Enable == 1 {
			enable = true
		}
		data.Rules[i].Enable = types.BoolValue(enable)
		data.Rules[i].ICMPType = types.StringValue(remoteRule.ICMPType)
		data.Rules[i].Interface = types.StringValue(remoteRule.Interface)
		data.Rules[i].IPVersion = types.Int64Value(int64(remoteRule.IPVersion))
		data.Rules[i].LogLevel = types.StringValue(string(remoteRule.LogLevel))
		data.Rules[i].Macro = types.StringValue(remoteRule.Macro)
		// FIXME: POS NOT WORKING
		//data.Rules[i].Pos = types.Int64Value(int64(remoteRule.Pos))
		data.Rules[i].Proto = types.StringValue(remoteRule.Proto)
		data.Rules[i].Source = types.StringValue(remoteRule.Source)
		data.Rules[i].Sport = types.StringValue(remoteRule.Sport)
		data.Rules[i].Type = types.StringValue(remoteRule.Type)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RulesResourceModel
	var state RulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	for _, rule := range state.Rules {
		if err := r.client.Node.Firewall.DeleteRule(data.Node.ValueString(), rule.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
			return
		}
	}

	for i, rule := range data.Rules {
		enable := 0
		if rule.Enable.ValueBool() {
			enable = 1
		}

		apiReq := pve.CreateNodeFirewallRuleRequest{
			Action:      rule.Action.ValueString(),
			Node:        data.Node.ValueString(),
			Type:        rule.Type.ValueString(),
			Comment:     rule.Comment.ValueString(),
			Destination: rule.Destination.ValueString(),
			//Digest:          rule.Digest.ValueString(),
			DestinationPort: int(rule.DestinationPort.ValueInt64()),
			Enable:          enable,
			ICMPType:        rule.ICMPType.ValueString(),
			Interface:       rule.Interface.ValueString(),
			LogLevel:        pve.FirewallLogLevel(rule.LogLevel.ValueString()),
			Macro:           rule.Macro.ValueString(),
			// FIXME: POS NOT WORKING
			// Pos:             int(rule.Pos.ValueInt64()),
			Proto:  rule.Proto.ValueString(),
			Source: rule.Source.ValueString(),
			Sport:  rule.Sport.ValueString(),
		}

		id, err := r.client.Node.Firewall.NewRule(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
			return
		}
		data.Rules[i].ID = types.StringValue(id)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RulesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	for _, rule := range data.Rules {
		if err := r.client.Node.Firewall.DeleteRule(data.Node.ValueString(), rule.ID.ValueString()); err != nil {
			tflog.Error(ctx, err.Error())
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node firewall rules, got error: %s", err))
			return
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "deleted a resource")
}

func (r *RulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
