package lxc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
var _ resource.Resource = &LXCResource{}
var _ resource.ResourceWithImportState = &LXCResource{}

func NewLXCResource(name string) func() resource.Resource {
	return func() resource.Resource {
		return &LXCResource{name: name}
	}
}

// LXCResource defines the resource implementation.
type LXCResource struct {
	client *pve.PVE
	name   string
}

type LXCFeaturesResourceModel struct {
	ForceRWSys types.Bool `tfsdk:"force_rw_sys"`
	Fuse       types.Bool `tfsdk:"fuse"`
	KeyCTL     types.Bool `tfsdk:"key_ctl"`
	Nesting    types.Bool `tfsdk:"nesting"`
	// TODO: Add support for mknod.
	// mknod types.idk `tfsdk:"mknod"`?
}

func (m *LXCFeaturesResourceModel) LoadFromObject(ctx context.Context, obj types.Object) {
	obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
		UnhandledNullAsEmpty:    true,
	})
}

func (m LXCFeaturesResourceModel) ToPVELXCFeatures() pve.LXCFeatures {
	f := pve.LXCFeatures{}

	f.ForceRWSys = m.ForceRWSys.ValueBoolPointer()
	f.Fuse = m.Fuse.ValueBoolPointer()
	f.KeyCTL = m.KeyCTL.ValueBoolPointer()
	f.Nesting = m.Nesting.ValueBoolPointer()

	return f
}

type LXCRootFSResourceModel struct {
	Volume types.String `tfsdk:"volume"`
	ACL    types.Bool   `tfsdk:"acl"`
	// TODO: Add support for mountoptions.
	Quota     types.Bool  `tfsdk:"quota"`
	Replicate types.Bool  `tfsdk:"replicate"`
	ReadOnly  types.Bool  `tfsdk:"read_only"`
	Shared    types.Bool  `tfsdk:"shared"`
	DiskSize  types.Int64 `tfsdk:"disk_size"`
}

type LXCNetResourceModel struct {
	Name       string  `tfsdk:"name"`
	Bridge     *string `tfsdk:"bridge"`
	Firewall   *bool   `tfsdk:"firewall"`
	GW         *string `tfsdk:"gateway"`
	GW6        *string `tfsdk:"gateway6"`
	HWAddr     *string `tfsdk:"hw_address"`
	IP         *string `tfsdk:"ip"`
	IP6        *string `tfsdk:"ip6"`
	LinkDown   *bool   `tfsdk:"link_down"`
	MTU        *int    `tfsdk:"mtu"`
	Rate       *int    `tfsdk:"rate"`
	Tag        *int    `tfsdk:"tag"`
	ComputedIP *string `tfsdk:"computed_ip"`
}

func (m *LXCNetResourceModel) LoadFromObject(ctx context.Context, obj types.Object) {
	obj.As(ctx, m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
}

func (m LXCNetResourceModel) ToObject() types.Object {
	elementTypes := map[string]attr.Type{
		"name":        types.StringType,
		"bridge":      types.StringType,
		"firewall":    types.BoolType,
		"gateway":     types.StringType,
		"gateway6":    types.StringType,
		"hw_address":  types.StringType,
		"ip":          types.StringType,
		"ip6":         types.StringType,
		"link_down":   types.BoolType,
		"mtu":         types.Int64Type,
		"rate":        types.Int64Type,
		"tag":         types.Int64Type,
		"computed_ip": types.StringType,
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
	if m.ComputedIP != nil {
		elements["computed_ip"] = types.StringValue(*m.ComputedIP)
	}
	object, _ := types.ObjectValueFrom(context.TODO(), elementTypes, m)

	return object
}

func (m LXCNetResourceModel) ToPVELXCNet() pve.LxcNet {
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

// LXCResourceModel describes the resource data model.
type LXCResourceModel struct {
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
	Status types.String   `tfsdk:"status"`
	CMDs   []types.String `tfsdk:"cmds"`
}

func (r *LXCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, r.name)
}

func (r *LXCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "LXC resource",
		Description:         DESC_LXC,
		Attributes:          newLXCResourceAttrs(),
		Blocks: map[string]schema.Block{
			"root_fs": schema.SingleNestedBlock{
				Description: DESC_LXC_ROOTFS,
				Attributes:  newLXCRootFSResourceAttrs(),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"features": schema.SingleNestedBlock{
				Description: DESC_LXC_FEATS,
				Attributes:  newLXCFeaturesResourceAttrs(),
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

func (r *LXCResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LXCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LXCResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if no vmid has been set, retrieve one from proxmox
	vmid, err := getVMID(r.client, data.VMID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
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
	apiReq.Net = newPVELXCNets(ctx, data.Networks)
	tflog.Debug(ctx, "got networks", map[string]any{"networks": apiReq.Net})

	// send lxc create request through api
	if _, err := r.client.LXC.Create(apiReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err.Error()))
		return
	} else {
		tflog.Info(ctx, "lxc created", map[string]any{"vmid": vmid, "req": apiReq})
	}

	// appends the current state after creation
	data.VMID = types.Int64Value(int64(vmid))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Start or stop the lxc according to the configured status
	err = updateLXCStatus(
		ctx,
		r.client,
		apiReq.Node,
		vmid,
		data.Status.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err.Error()))
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

	// Appends the current state after creation
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Compute network ips only if lxc is running.
	// At this point, we know the lxc desired status.
	if data.Status.ValueString() == string(pve.LXC_STATUS_RUNNING) {
		computedNets, err := computeLXCNetIPs(ctx, r.client, apiReq.Node, vmid, data.Networks)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err.Error()))
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

		tflog.Info(ctx, "proxmox_lxc_read_computed_ips", map[string]any{"networks": newLXCNetsResourceModel(ctx, computedNets)})
		data.Networks = computedNets
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// run commands in lxc
	// if desired status is running, simply run the commands
	// if desired status is stopped, start -> run cmds -> stop
	if len(data.CMDs) > 0 {
		// if the desiredStatus is stopped start the lxc
		if err := updateLXCStatus(
			ctx,
			r.client,
			apiReq.Node,
			vmid,
			string(pve.LXC_STATUS_RUNNING),
		); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start node lxc, got error: %s", err))
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

		if err := updateLXCStatus(
			ctx,
			r.client,
			apiReq.Node,
			vmid,
			data.Status.ValueString(),
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

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: Use lxc endpoints to read mem, cpu and so on
func (r *LXCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LXCResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmid := int(data.VMID.ValueInt64())
	node := data.Node.ValueString()

	desiredStatus := data.Status.ValueString()

	remoteData, err := r.client.LXC.GetByID(node, vmid)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node lxc, got error: %s", err))
		return
	} else if remoteData == nil {
		msg := fmt.Sprintf("LXC %d not found in node %s, maybe it was deleted. Try removing it manually from the state", vmid, node)
		resp.Diagnostics.AddError("Client Error", msg)
		return
	}

	data.Status = types.StringValue(string(remoteData.Status))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if desiredStatus == string(pve.LXC_STATUS_RUNNING) {
		computedNets, err := computeLXCNetIPs(ctx, r.client, node, vmid, data.Networks)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read lxc ifaces ips, got error: %s", err.Error()))
		}

		tflog.Info(ctx, "proxmox_lxc_read_computed_ips", map[string]any{"networks": newLXCNetsResourceModel(ctx, computedNets)})
		data.Networks = computedNets
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: while lxc endpoints are implemented and not every property
// requires a replace, only do a start/stop.
func (r *LXCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LXCResourceModel
	var state LXCResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	node := state.Node.ValueString()
	vmid := int(state.VMID.ValueInt64())
	status := state.Status.ValueString()

	tflog.Info(ctx, "proxmox_lxc_update_started", map[string]any{"node": node, "vmid": vmid})

	if resp.Diagnostics.HasError() {
		return
	}

	if err := updateLXCStatus(ctx, r.client, node, vmid, string(pve.LXC_STATUS_STOPPED)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop lxc, got error: %s", err))
		return
	}

	tflog.Info(ctx, "proxmox_lxc_update_networks", map[string]any{
		"node":          node,
		"vmid":          vmid,
		"planNetworks":  newLXCNetsResourceModel(ctx, plan.Networks),
		"stateNetworks": newLXCNetsResourceModel(ctx, state.Networks),
	})
	if err := r.client.LXC.Update(pve.UpdateLxcRequest{
		Node: state.Node.ValueString(),
		VMID: vmid,
		Net:  newPVELXCNets(ctx, plan.Networks),
	}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update lxc interfaces, got error: %s", err))
		return
	}
	if err := updateLXCStatus(ctx, r.client, node, vmid, status); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update lxc status, got error: %s", err))
		return
	}
	computedNets, err := computeLXCNetIPs(ctx, r.client, node, vmid, plan.Networks)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read lxc ifaces ips, got error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "proxmox_lxc_update_computed_ips", map[string]any{
		"node":     node,
		"vmid":     vmid,
		"networks": newLXCNetsResourceModel(ctx, computedNets),
	})
	state.Networks = computedNets

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LXCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LXCResourceModel

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

	tflog.Trace(ctx, "deleted an lxc resource")
}

// TODO: implement
func (r *LXCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
