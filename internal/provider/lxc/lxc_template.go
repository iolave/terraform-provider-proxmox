package lxc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LXCTplResource{}
var _ resource.ResourceWithImportState = &LXCTplResource{}

func NewLXCTplResource(name string) func() resource.Resource {
	return func() resource.Resource {
		return &LXCTplResource{name: name}
	}
}

// LXCTplResource defines the resource implementation.
type LXCTplResource struct {
	client *pve.PVE
	name   string
}

// LXCTplResourceModel describes the resource data model.
type LXCTplResourceModel struct {
	// Create options
	Node       types.String `tfsdk:"node"`
	OSTemplate types.String `tfsdk:"os_template"`
	VMID       types.Int64  `tfsdk:"id"`
	//Arch        types.String `tfsdk:"arch"`
	//BWLimit types.Int64 `tfsdk:"bwlimit"`
	//ConsoleMode types.String `tfsdk:"console_mode"`
	//Console     types.Bool   `tfsdk:"console"`
	//Cores       types.Int64  `tfsdk:"cores"`
	//CPULimit    types.Int64  `tfsdk:"cpu_limit"`
	//CPUUnits    types.Int64  `tfsdk:"cpu_units"`
	//Debug       types.Bool   `tfsdk:"debug"`
	//Description types.String `tfsdk:"description"`
	// TODO: Add support for devices. "dev[n]" in the docs.
	// dev[n] types.idk

	Features types.Object `tfsdk:"features"`
	//Force              types.Bool   `tfsdk:"force"`
	//Hookscript         types.String `tfsdk:"hookscript"`
	Hostname types.String `tfsdk:"hostname"`
	//IgnoreUnpackErrors types.Bool   `tfsdk:"ignore_unpack_errors"`
	//Lock   types.String `tfsdk:"lock"`
	//Memory types.Int64 `tfsdk:"memory"`
	// TODO: Add support for mount points. "mp[n]" in the docs.
	// mp[n] types.idk

	Nameserver types.String `tfsdk:"nameserver"`
	//TODO: Add support for multiple networks. "net[n]" in the docs.
	Networks []types.Object `tfsdk:"networks"`
	OnBoot   types.Bool     `tfsdk:"on_boot"`
	//OSType        types.String   `tfsdk:"os_type"`
	Password types.String `tfsdk:"password"`
	//Pool          types.String   `tfsdk:"pool"`
	//Protection    types.Bool     `tfsdk:"protection"`
	//Restore       types.Bool     `tfsdk:"restore"`
	RootFS types.Object `tfsdk:"root_fs"`
	//SearchDomain  types.String   `tfsdk:"search_domain"`
	SSHPublicKeys []types.String `tfsdk:"ssh_public_keys"`
	//Start         types.Bool     `tfsdk:"start"`
	//TODO: Add support for startup

	// Startup LXCStartupResourceModel `tfsdk:"startup"`
	//StorageID    types.String   `tfsdk:"storage_id"`
	//SwapSize     types.Int64    `tfsdk:"swap_size"`
	//Tags         []types.String `tfsdk:"tags"`
	//Template     types.Bool     `tfsdk:"template"`
	//Timezone     types.String   `tfsdk:"timezone"`
	//AvailableTTY types.Int64    `tfsdk:"available_tty"`
	//UniqueHWAddr types.Bool     `tfsdk:"unique_hw_addr"`
	Unprivileged types.Bool `tfsdk:"unprivileged"`
	// TODO: Add support for unused[n]?

	// Custom
	CMDs []types.String `tfsdk:"cmds"`
}

func (r *LXCTplResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, r.name)
}

func (r *LXCTplResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "LXC Template",
		Description:         DESC_LXC,
		Attributes:          newLXCTplResourceAttrs(),
		Blocks: map[string]schema.Block{
			"root_fs": schema.SingleNestedBlock{
				Description: DESC_LXC_ROOTFS,
				Attributes:  newLxcTplRootFSResourceAttrs(),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"features": schema.SingleNestedBlock{
				Description: DESC_LXC_FEATS,
				Attributes:  newLxcTplFeaturesResourceAttrs(),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}

	if r.name == "node_lxc" {
		resp.Schema.DeprecationMessage = "Use proxmox_lxc resource instead. This resource will be removed in the next major version of the provider."
	}
}

func (r *LXCTplResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LXCTplResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LXCTplResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if no vmid has been set, retrieve one from proxmox
	vmid, err := getVMID(r.client, data.VMID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %#v", err))
		return
	} else {
		tflog.Info(ctx, "got vmid from proxmox", map[string]any{"vmid": vmid})
	}

	// Format the ssh public keys to the go-proxmox format
	ssh := formatSSHPublicKey(data.SSHPublicKeys)

	// Create resource
	apiReq := pve.CreateLxcRequest{
		Node:          data.Node.ValueString(),
		OSTemplate:    data.OSTemplate.ValueString(),
		VMID:          vmid,
		Hostname:      data.Hostname.ValueString(),
		Password:      data.Password.ValueString(),
		SSHPublicKeys: ssh,
	}

	if unpriv := data.Unprivileged.ValueBoolPointer(); unpriv != nil {
		apiReq.Unprivileged = *unpriv
	}

	if onBoot := data.OnBoot.ValueBoolPointer(); onBoot != nil {
		apiReq.OnBoot = *onBoot
	}
	if onBoot := data.OnBoot.ValueBoolPointer(); onBoot != nil {
		apiReq.OnBoot = *onBoot
	}
	if ns := data.Nameserver.ValueStringPointer(); ns != nil {
		apiReq.Nameserver = *ns
	}

	// Set features to api request
	apiReq.Features = newLXCFeaturesResourceModel(ctx, data.Features).
		ToPVELXCFeatures()
	tflog.Debug(ctx, "got features", map[string]any{"features": apiReq.Features})

	// set networks to api request
	apiReq.Net = newPVELXCTplNets(ctx, data.Networks)

	// send lxc create request through api
	if _, err := r.client.LXC.Create(apiReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %v", err))
		return
	} else {
		tflog.Info(ctx, "lxc created", map[string]any{"vmid": vmid, "req": apiReq})
	}

	// appends the current state after creation
	data.VMID = types.Int64Value(int64(vmid))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// run commands in lxc
	// if desired status is running, simply run the commands
	// if desired status is stopped, start -> run cmds -> stop
	if len(data.CMDs) > 0 {
		// start the lxc to run the commands
		if err := updateLXCStatus(
			ctx,
			r.client,
			apiReq.Node,
			vmid,
			string(pve.LXC_STATUS_RUNNING),
		); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start lxc to run commands, got error: %s", err))
			if err := deleteLXC(
				ctx,
				r.client,
				apiReq.Node,
				vmid,
			); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete lxc, got error: %s", err.Error()))
			} else {
				resp.State.RemoveResource(ctx)
			}
			return
		}

		// run commands
		if err := runLXCCommands(ctx, r.client, vmid, data.CMDs); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to run commands inside lxc , got error: %s", err))
			if err := deleteLXC(
				ctx,
				r.client,
				apiReq.Node,
				vmid,
			); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete lxc, got error: %s", err.Error()))
			} else {
				resp.State.RemoveResource(ctx)
			}
			return
		}

		// stop the lxc after commands are run
		if err := updateLXCStatus(
			ctx,
			r.client,
			apiReq.Node,
			vmid,
			string(pve.LXC_STATUS_STOPPED),
		); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node lxc status, got error: %s", err))
			if err := deleteLXC(
				ctx,
				r.client,
				apiReq.Node,
				vmid,
			); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete lxc, got error: %s", err.Error()))
			} else {
				resp.State.RemoveResource(ctx)
			}
			return
		}
	}

	// Convert the lxc to a template
	if err := r.client.LXC.CreateTemplate(apiReq.Node, vmid); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert lxc to template, got error: %s", err.Error()))
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: Use lxc endpoints to read mem, cpu and so on
func (r *LXCTplResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LXCTplResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: implement update content on changes

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LXCTplResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *LXCTplResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LXCTplResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteLXC(
		ctx,
		r.client,
		data.Node.ValueString(),
		int(data.VMID.ValueInt64()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
		return
	}
}

func (r *LXCTplResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import lxc template, got error: %s", err))
		return
	}

	state := LXCTplResourceModel{
		VMID: basetypes.NewInt64Value(int64(id)),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type LXCTplNetResourceModel struct {
	Name     string  `tfsdk:"name"`
	Bridge   *string `tfsdk:"bridge"`
	Firewall *bool   `tfsdk:"firewall"`
	GW       *string `tfsdk:"gateway"`
	GW6      *string `tfsdk:"gateway6"`
	HWAddr   *string `tfsdk:"hw_address"`
	IP       *string `tfsdk:"ip"`
	IP6      *string `tfsdk:"ip6"`
	LinkDown *bool   `tfsdk:"link_down"`
	MTU      *int    `tfsdk:"mtu"`
	Rate     *int    `tfsdk:"rate"`
	Tag      *int    `tfsdk:"tag"`
}

func (m *LXCTplNetResourceModel) LoadFromObject(ctx context.Context, obj types.Object) {
	obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
}

func (m LXCTplNetResourceModel) ToObject() types.Object {
	elementTypes := map[string]attr.Type{
		"name":       types.StringType,
		"bridge":     types.StringType,
		"firewall":   types.BoolType,
		"gateway":    types.StringType,
		"gateway6":   types.StringType,
		"hw_address": types.StringType,
		"ip":         types.StringType,
		"ip6":        types.StringType,
		"link_down":  types.BoolType,
		"mtu":        types.Int64Type,
		"rate":       types.Int64Type,
		"tag":        types.Int64Type,
	}
	elements := map[string]attr.Value{}
	elements["name"] = types.StringValue(m.Name)
	elements["bridge"] = types.StringPointerValue(m.Bridge)
	elements["firewall"] = types.BoolPointerValue(m.Firewall)
	elements["gateway"] = types.StringPointerValue(m.GW)
	elements["gateway6"] = types.StringPointerValue(m.GW6)
	elements["hw_address"] = types.StringPointerValue(m.HWAddr)
	elements["ip"] = types.StringPointerValue(m.IP)
	elements["ip6"] = types.StringPointerValue(m.IP6)
	elements["link_down"] = types.BoolPointerValue(m.LinkDown)
	if m.MTU != nil {
		elements["mtu"] = types.Int64Value(int64(*m.MTU))
	}
	if m.Rate != nil {
		elements["rate"] = types.Int64Value(int64(*m.Rate))
	}
	if m.Tag != nil {
		elements["tag"] = types.Int64Value(int64(*m.Tag))
	}
	object, _ := types.ObjectValueFrom(context.TODO(), elementTypes, m)

	return object
}

func (m LXCTplNetResourceModel) ToPVELXCNet() pve.LxcNet {
	pveNet := pve.LxcNet{Name: m.Name}
	if m.Bridge != nil {
		pveNet.Bridge = *m.Bridge
	}
	if m.Firewall != nil {
		pveNet.Firewall = *m.Firewall
	}
	if m.IP != nil {
		pveNet.IP = *m.IP
	}
	if m.IP6 != nil {
		pveNet.IP6 = *m.IP6
	}
	if m.HWAddr != nil {
		pveNet.HWAddr = *m.HWAddr
	}
	if m.GW != nil {
		pveNet.GW = *m.GW
	}
	if m.GW6 != nil {
		pveNet.GW6 = *m.GW6
	}
	if m.LinkDown != nil {
		pveNet.LinkDown = *m.LinkDown
	}
	if m.MTU != nil {
		pveNet.MTU = *m.MTU
	}
	if m.Rate != nil {
		pveNet.Rate = *m.Rate
	}
	if m.Tag != nil {
		pveNet.Tag = *m.Tag
	}
	return pveNet
}
