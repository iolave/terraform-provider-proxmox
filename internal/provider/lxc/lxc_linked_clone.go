package lxc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LXCLinkedCloneResource{}
var _ resource.ResourceWithImportState = &LXCLinkedCloneResource{}

func NewLXCLinkedCloneResource() resource.Resource {
	return &LXCLinkedCloneResource{}
}

// LXCLinkedCloneResource defines the resource implementation.
type LXCLinkedCloneResource struct {
	client *pve.PVE
}

// LXCLinkedCloneResourceModel describes the resource data model.
type LXCLinkedCloneResourceModel struct {
	VMID     types.Int64  `tfsdk:"source_id"`
	Node     types.String `tfsdk:"node"`
	NewVMID  types.Int64  `tfsdk:"id"`
	BWLimit  types.Int64  `tfsdk:"bwlimit"`
	Desc     types.String `tfsdk:"description"`
	Hostname types.String `tfsdk:"hostname"`
	Pool     types.String `tfsdk:"pool"`
	Snapname types.String `tfsdk:"snapshot_name"`
	Status   types.String `tfsdk:"status"`
	// Target node. Only allowed if the original VM is on shared storage.
	// TODO: Target types.String `tfsdk:"target"`

	// READ ONLY PROPERTIES
	Networks types.List `tfsdk:"networks"`
}

func (r *LXCLinkedCloneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	name := "lxc_linked_clone"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (r *LXCLinkedCloneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Do a linked clone of an lxc",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.Int64Attribute{
				Description: DESC_LXC_ID,
				Required:    true,
			},
			"node": schema.StringAttribute{
				Description: DESC_LXC_NODE,
				Required:    true,
			},
			"id": schema.Int64Attribute{
				Description: DESC_LXC_ID,
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"bwlimit": schema.Int64Attribute{
				Description: DESC_LXC_BWLIM,
				Optional:    true,
				//Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: DESC_LXC_DESC,
				Optional:    true,
				//Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: DESC_LXC_HOSTNAME,
				Optional:    true,
				//Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pool": schema.StringAttribute{
				Description: DESC_LXC_POOL,
				Optional:    true,
				//Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_name": schema.StringAttribute{
				Description: "The name of the snapshot to clone from",
				Optional:    true,
				//Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			//"target": schema.StringAttribute{
			//	Description: "The target node. Only allowed if the original VM is on shared storage.",
			//	Optional:    true,
			//	//Computed:    true,
			//	PlanModifiers: []planmodifier.String{
			//		stringplanmodifier.RequiresReplace(),
			//	},
			//},
			"status": schema.StringAttribute{
				Description: DESC_LXC_STATUS,
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(pve.LXC_STATUS_STOPPED)),
			},

			// READ ONLY PROPERTIES
			"networks": schema.ListNestedAttribute{
				Description: "Computed ifaces",
				Computed:    true,
				//Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{
				//	AttrTypes: map[string]attr.Type{
				//		"name":  types.StringType,
				//		"ip_v4": types.StringType,
				//	},
				//}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: newLXCCloneNetResourceAttrs(),
				},
			},
			//"networks": schema.ListNestedAttribute{
			//	Description: "Computed ifaces",
			//	Computed:    true,
			//	//Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{
			//	//	AttrTypes: map[string]attr.Type{
			//	//		"name":  types.StringType,
			//	//		"ip_v4": types.StringType,
			//	//	},
			//	//}, []attr.Value{})),
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: newLXCCloneNetResourceAttrs(),
			//	},
			//},
		},
	}
}

func (r *LXCLinkedCloneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LXCLinkedCloneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LXCLinkedCloneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceId := int(data.VMID.ValueInt64())
	if sourceId == 0 {
		resp.Diagnostics.AddError("Client Error", "source_id property is required")
		return
	}

	node := data.Node.ValueString()
	if node == "" {
		resp.Diagnostics.AddError("Client Error", "node property is required")
		return
	}

	targetId := int(data.NewVMID.ValueInt64())
	if targetId == 0 {
		id, err := getVMID(r.client, data.NewVMID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to generate a vmid, got error: %s", err.Error()))
			return
		}
		targetId = id
	}

	apiReq := pve.CloneLxcRequest{
		Node:    node,
		VMID:    sourceId,
		NewVMID: targetId,
	}

	if data.BWLimit.ValueInt64Pointer() != nil {
		apiReq.BWLimit = int(data.BWLimit.ValueInt64())
	}
	if data.Desc.ValueStringPointer() != nil {
		apiReq.Desc = data.Desc.ValueString()
	}
	if data.Hostname.ValueStringPointer() != nil {
		apiReq.Hostname = data.Hostname.ValueString()
	}
	if data.Pool.ValueStringPointer() != nil {
		apiReq.Pool = data.Pool.ValueString()
	}
	if data.Snapname.ValueStringPointer() != nil {
		apiReq.Snapname = data.Snapname.ValueString()
	}
	tflog.Info(ctx, "proxmox_lxc_linked_clone_create_request", map[string]any{
		"request": apiReq,
	})

	if _, err := r.client.LXC.Clone(apiReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clone lxc, got error: %s", err.Error()))
		return
	}
	data.NewVMID = types.Int64Value(int64(targetId))
	// We set the networks to empty because we don't know yet
	// if the clone is running or not.
	data.Networks = types.ListNull(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"ip_v4": types.StringType,
		},
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	status := data.Status.ValueString()
	if err := updateLXCStatus(ctx, r.client, node, targetId, status); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update lxc status, got error: %s", err.Error()))
		if err := deleteLXC(
			ctx,
			r.client,
			node,
			targetId,
		); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete lxc, got error: %s", err.Error()))
		} else {
			resp.State.RemoveResource(ctx)
		}
		return
	}

	// If the clone is running, we compute the networks
	if status == string(pve.LXC_STATUS_RUNNING) {
		computedNets, err := computeLXCCloneNetIPs(
			r.client,
			node,
			targetId,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read lxc ifaces ips, got error: %s", err.Error()))
			return
		}
		data.Networks = computedNets
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LXCLinkedCloneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LXCLinkedCloneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(data.NewVMID.ValueInt64())
	node := data.Node.ValueString()

	desiredStatus := data.Status.ValueString()

	remoteData, err := r.client.LXC.GetByID(node, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node lxc, got error: %s", err))
		return
	} else if remoteData == nil {
		req.State.RemoveResource(ctx)
		msg := fmt.Sprintf("LXC %d not found in node %s, maybe it was deleted. It was removed from the state", id, node)
		tflog.Error(ctx, msg)
		return
	}

	data.Status = types.StringValue(string(remoteData.Status))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// If the clone is running, we compute the networks
	if desiredStatus == string(pve.LXC_STATUS_RUNNING) {
		computedNets, err := computeLXCCloneNetIPs(
			r.client,
			node,
			id,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read lxc ifaces ips, got error: %s", err.Error()))
			return
		}
		data.Networks = computedNets
	} else {
		data.Networks = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"ip_v4": types.StringType,
			},
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LXCLinkedCloneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LXCLinkedCloneResourceModel
	var state LXCLinkedCloneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	node := state.Node.ValueString()
	id := int(state.NewVMID.ValueInt64())
	status := plan.Status.ValueString()

	tflog.Info(ctx, "proxmox_lxc_linked_clone_update_started", map[string]any{"node": node, "vmid": id, "desiredStatus": status})

	if resp.Diagnostics.HasError() {
		return
	}

	if err := updateLXCStatus(ctx, r.client, node, id, status); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update lxc status, got error: %s", err))
		return
	}
	state.Status = types.StringValue(status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	// If the clone is running, we compute the networks
	if status == string(pve.LXC_STATUS_RUNNING) {
		computedNets, err := computeLXCCloneNetIPs(
			r.client,
			node,
			id,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read lxc ifaces ips, got error: %s", err.Error()))
			return
		}
		state.Networks = computedNets
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}

}

func (r *LXCLinkedCloneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LXCLinkedCloneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteLXC(
		ctx,
		r.client,
		data.Node.ValueString(),
		int(data.NewVMID.ValueInt64()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete lxc, got error: %s", err))
		return
	}
}

func (r *LXCLinkedCloneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import lxc linked clone, got error: %s", err))
		return
	}

	state := LXCLinkedCloneResourceModel{
		VMID: basetypes.NewInt64Value(int64(id)),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type LXCCloneNetResourceModel struct {
	Name *string `tfsdk:"name"`
	IPV4 *string `tfsdk:"ip_v4"`
}

func (m *LXCCloneNetResourceModel) LoadFromObject(ctx context.Context, obj types.Object) {
	obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
}

func (m *LXCCloneNetResourceModel) LoadFromPVE(net pve.GetLxcInterfaceResponse) {
	m.Name = &net.Name
	m.IPV4 = &net.IPv4
}

func (m LXCCloneNetResourceModel) ToObject() types.Object {
	elementTypes := map[string]attr.Type{
		"name":  types.StringType,
		"ip_v4": types.StringType,
	}

	elements := map[string]attr.Value{}
	if m.Name != nil {
		elements["name"] = types.StringValue(*m.Name)
	}
	if m.IPV4 != nil {
		elements["ip_v4"] = types.StringValue(*m.IPV4)
	}

	object, _ := types.ObjectValueFrom(context.TODO(), elementTypes, m)

	return object
}

func newLXCCloneNetResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed: true,
		},
		"ip_v4": schema.StringAttribute{
			Description: DESC_LXC_IP,
			Computed:    true,
		},
	}
}
