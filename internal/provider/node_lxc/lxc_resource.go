package nodelxc

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func NewLXCResource() resource.Resource {
	return &LXCResource{}
}

// LXCResource defines the resource implementation.
type LXCResource struct {
	client *pve.PVE
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

	//Nameserver types.String `tfsdk:"nameserver"`
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
	name := "node_lxc"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (r *LXCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Node lxc resource",
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
	var vmid int
	var err error
	if data.VMID.IsNull() || data.VMID.IsUnknown() {
		vmid, err = r.client.Cluster.GetRandomVMID()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
			return
		}
	} else {
		vmid = int(data.VMID.ValueInt64())
	}
	tflog.Info(ctx, "got vmid from proxmox", map[string]any{"id": vmid})

	// marshals ssh keys
	ssh := ""
	for _, pub := range data.SSHPublicKeys {
		ssh = fmt.Sprintf("%s\n%s", ssh, pub.ValueString())
	}

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

	// Set features to api request
	feats := LXCFeaturesResourceModel{}
	feats.LoadFromObject(ctx, data.Features)
	tflog.Debug(ctx, "got features", map[string]any{"features": feats.ToPVELXCFeatures()})
	apiReq.Features = feats.ToPVELXCFeatures()

	// set networks to api request
	networksCopy := []types.Object{}
	for _, v := range data.Networks {
		networksCopy = append(networksCopy, v)
	}

	for i, obj := range networksCopy {
		net := LXCNetResourceModel{}
		net.LoadFromObject(ctx, obj)
		apiReq.Net = append(apiReq.Net, net.ToPVELXCNet())
		tflog.Debug(ctx, "configured network", map[string]any{
			"pos":     i,
			"network": net,
		})
	}

	// send lxc create request through api
	createRes, err := r.client.LXC.Create(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err.Error()))
		return
	}
	tflog.Info(ctx, "lxc created", map[string]any{"id": createRes, "req": apiReq})
	// sets vmid to the state
	data.VMID = types.Int64Value(int64(createRes))

	// Start or stop the lxc according to the configured status
	desiredStatus := data.Status.ValueString()
	for true {
		time.Sleep(time.Second * 8)

		tflog.Info(ctx, "querying lxc status", map[string]any{"id": createRes, "desired": desiredStatus})
		remoteStatus, err := r.client.LXC.GetStatus(
			data.Node.ValueString(),
			vmid,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
			return
		}
		tflog.Info(ctx, "got lxc status", map[string]any{"id": createRes, "desired": desiredStatus, "remote": remoteStatus.Status})

		if desiredStatus == remoteStatus.Status {
			break
		}

		switch desiredStatus {
		case string(pve.LXC_STATUS_RUNNING):
			if remoteStatus.Status != string(pve.LXC_STATUS_STOPPED) {
				continue
			}
			_, err = r.client.LXC.Start(pve.LXCStartRequest{Node: data.Node.ValueString(), ID: vmid})

		case string(pve.LXC_STATUS_STOPPED):
			if remoteStatus.Status != string(pve.LXC_STATUS_RUNNING) {
				continue
			}
			_, err = r.client.LXC.Stop(pve.LXCStopRequest{Node: data.Node.ValueString(), ID: vmid})
		}
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
			return
		}
	}

	// Appends the current state after creation
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Compute network ips only if lxc is running.
	// At this point, we know the lxc desired status.
	if desiredStatus == string(pve.LXC_STATUS_RUNNING) {
		ipRetries := 3
		ipSleepTime := time.Second * 15
		for i := 0; i < ipRetries; i++ {
			time.Sleep(ipSleepTime)

			// retrieve lxc interfaces of running lxc
			// and store them in the map below for easy
			// access through the iface name.
			ifacesMap := map[string]pve.GetLxcInterfaceResponse{}
			ifaces, err := r.client.LXC.GetInterfaces(apiReq.Node, vmid)
			if err != nil {
				tflog.Error(ctx, "failed to retrieve lxc ifaces", map[string]interface{}{"error": err.Error(), "try": i})
				continue
			}
			for _, iface := range ifaces {
				tflog.Debug(ctx, "got interface", map[string]interface{}{"iface": iface, "try": i})
				ifacesMap[iface.Name] = iface
			}

			// Read configured networks
			computedNets := map[string]LXCNetResourceModel{}
			for _, obj := range networksCopy {
				net := LXCNetResourceModel{}
				net.LoadFromObject(ctx, obj)

				tflog.Debug(ctx, "mapping network to set computed ip", map[string]interface{}{"try": i, "network": net})

				if ifacesMap[net.Name].IPv4 == "" {
					tflog.Error(ctx, "lxc iface does not have an assigned ip", map[string]interface{}{"try": i, "net": net, "iface": ifacesMap[net.Name]})
					break
				}

				ip := ifacesMap[net.Name].IPv4
				net.ComputedIP = &ip
				computedNets[net.Name] = net
			}

			tflog.Debug(ctx, "network slice and map sizes", map[string]interface{}{
				"try":              i,
				"computedNets":     len(computedNets),
				"data.Networks":    len(data.Networks),
				"computedNetsData": computedNets,
			})
			if len(data.Networks) != len(computedNets) {
				err := errors.New("Unable to compute all ifaces ips")
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
				return
			}

			data.Networks = []types.Object{}
			for i, netPosObj := range networksCopy {
				netPosModel := LXCNetResourceModel{}
				netPosModel.LoadFromObject(ctx, netPosObj)
				net := computedNets[netPosModel.Name]
				tflog.Debug(ctx, "got network with computed ip", map[string]interface{}{"try": i, "network": net, "pos": i})

				data.Networks = append(data.Networks, net.ToObject())
			}

			for i, obj := range data.Networks {
				net := LXCNetResourceModel{}
				net.LoadFromObject(ctx, obj)
				tflog.Debug(ctx, "updated data network", map[string]interface{}{"try": i, "network": net, "pos": i})
			}
		}
	}

	for _, v := range data.Networks {
		ip := v.Attributes()["computed_ip"].String()
		tflog.Debug(ctx, "data network", map[string]interface{}{"ip": ip})
	}

	tflog.Debug(ctx, "data after computed ips", map[string]interface{}{"data": data})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// run commands in lxc
	// if desired status is running, simply run the commands
	// if desired status is stopped, start -> run cmds -> stop
	if len(data.CMDs) > 0 {
		// if the desiredStatus is stopped start the lxc
		if desiredStatus == string(pve.LXC_STATUS_STOPPED) {
			tflog.Info(ctx, "starting lxc to run comands", map[string]any{"id": createRes})

			for true {
				time.Sleep(time.Second * 8)

				remoteStatus, err := r.client.LXC.GetStatus(
					data.Node.ValueString(),
					vmid,
				)
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
					return
				}

				data.Status = types.StringValue(remoteStatus.Status)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				tflog.Debug(ctx, "got lxc status while starting to tun commands", map[string]any{"id": createRes, "remote": remoteStatus.Status})

				if string(pve.LXC_STATUS_RUNNING) == remoteStatus.Status {
					break
				}
			}
		}

		// run commands
		for _, cmd := range data.CMDs {
			cmd := cmd.ValueString()
			// TODO: add support for more shells
			out, exit, err := r.client.LXC.Exec(vmid, "bash", cmd)
			tflog.Info(ctx, "executed cmd", map[string]any{"cmd": cmd, "output": out, "exitCode": exit, "error": err})
			if exit != 0 {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc cuz command execution failed, got error: %s", err.Error()))
				return
			}

			if exit == 0 {
				continue
			}
			// fail on command non zero exit code
			err = fmt.Errorf(`failed to execute command "%s", exit_code=%d, output=%s`, cmd, exit, out)
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
			return
		}

		// if the desiredStatus is stopped stop the lxc
		if desiredStatus == string(pve.LXC_STATUS_STOPPED) {
			tflog.Info(ctx, "stopping lxc after commands executed", map[string]any{"id": createRes})

			for true {
				time.Sleep(time.Second * 8)

				remoteStatus, err := r.client.LXC.GetStatus(
					data.Node.ValueString(),
					vmid,
				)
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
					return
				}
				data.Status = types.StringValue(remoteStatus.Status)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				tflog.Debug(ctx, "got lxc status while stopping after command ran", map[string]any{"id": createRes, "remote": remoteStatus.Status})

				if string(pve.LXC_STATUS_STOPPED) == remoteStatus.Status {
					break
				}
			}
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

	desiredStatus := data.Status.ValueString()

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	remoteData, err := r.client.LXC.GetAll(
		data.Node.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node lxc, got error: %s", err))
		return
	}

	status := ""
	for _, lxc := range remoteData {
		if lxc.VMID != int(data.VMID.ValueInt64()) {
			continue
		}
		status = string(lxc.Status)
		break
	}

	if status == "" {
		err := fmt.Errorf("node lxc with id %d not found", data.VMID.ValueInt64())
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node lxc, got error: %s", err))
		return
	}
	data.Status = types.StringValue(status)

	networksCopy := make([]types.Object, len(data.Networks))
	copy(networksCopy, data.Networks)
	if desiredStatus == string(pve.LXC_STATUS_RUNNING) {
		ipRetries := 3
		ipSleepTime := time.Second * 15
		for i := 0; i < ipRetries; i++ {
			time.Sleep(ipSleepTime)
			computedNets := []LXCNetResourceModel{}
			ifaces, err := r.client.LXC.GetInterfaces(data.Node.String(), int(data.VMID.ValueInt64()))
			if err != nil {
				tflog.Error(ctx, "failed to retrieve lxc ifaces", map[string]interface{}{"error": err.Error(), "try": i})
				continue
			}

			data.Networks = []types.Object{}
			for _, obj := range networksCopy {
				net := LXCNetResourceModel{}
				net.LoadFromObject(ctx, obj)

				var remoteIface pve.GetLxcInterfaceResponse
				for _, iface := range ifaces {
					if iface.Name != net.Name {
						continue
					}
					remoteIface = iface
				}
				if remoteIface.IPv4 == "" {
					tflog.Error(ctx, "lxc iface does not have an assigned ip", map[string]interface{}{"try": i})
					break
				}
				net.ComputedIP = &remoteIface.IPv4
				computedNets = append(computedNets, net)
			}
			if len(data.Networks) != len(computedNets) {
				err := errors.New("Unable to compute all ifaces ips")
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node lxc, got error: %s", err))
				return
			}

			data.Networks = []types.Object{}
			for i, netPosObj := range networksCopy {
				netPosModel := LXCNetResourceModel{}
				netPosModel.LoadFromObject(ctx, netPosObj)
				net := computedNets[i]
				tflog.Debug(ctx, "got network with computed ip", map[string]interface{}{"try": i, "network": net, "pos": i})

				data.Networks = append(data.Networks, net.ToObject())
			}

			for i, obj := range data.Networks {
				net := LXCNetResourceModel{}
				net.LoadFromObject(ctx, obj)
				tflog.Debug(ctx, "updated data network", map[string]interface{}{"try": i, "network": net, "pos": i})
			}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: while lxc endpoints are implemented and not every property
// requires a replace, only do a start/stop.
func (r *LXCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LXCResourceModel
	var state LXCResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	data.VMID = types.Int64Value(state.VMID.ValueInt64())

	if resp.Diagnostics.HasError() {
		return
	}

	// Update status if neccessary
	if data.Status.ValueString() != state.Status.ValueString() {
		var err error
		switch state.Status.ValueString() {
		case string(pve.LXC_STATUS_STOPPED):
			_, err = r.client.LXC.Start(pve.LXCStartRequest{
				Node: state.Node.ValueString(),
				ID:   int(state.VMID.ValueInt64()),
			})
		case string(pve.LXC_STATUS_RUNNING):
			_, err = r.client.LXC.Stop(pve.LXCStopRequest{
				Node:             state.Node.ValueString(),
				ID:               int(state.VMID.ValueInt64()),
				OverruleShutdown: 1,
			})
		}

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node lxc, got error: %s", err))
		}

		// check status in proxmox
		for true {
			time.Sleep(time.Second * 2)
			remoteStatus, err := r.client.LXC.GetStatus(
				state.Node.ValueString(),
				int(state.VMID.ValueInt64()),
			)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node lxc, got error: %s", err))
				return
			}
			if remoteStatus.Status == state.Status.ValueString() {
				continue
			}
			break
		}
	}

	//err := errors.New("unable to update cuz proxmox client have not implemented some features")
	//resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node lxc, got error: %s", err))
	//return

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LXCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LXCResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Stop the lxc if running
	remoteStatus, err := r.client.LXC.GetStatus(
		data.Node.ValueString(),
		int(data.VMID.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
		return
	}
	if remoteStatus.Status == string(pve.LXC_STATUS_RUNNING) {
		_, err = r.client.LXC.Stop(pve.LXCStopRequest{
			Node: data.Node.ValueString(),
			ID:   int(data.VMID.ValueInt64()),
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
			return
		}

	}
	// check if stopped
	for true {
		time.Sleep(time.Second * 2)
		remoteStatus, err := r.client.LXC.GetStatus(
			data.Node.ValueString(),
			int(data.VMID.ValueInt64()),
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
			return
		}
		if remoteStatus.Status != string(pve.LXC_STATUS_STOPPED) {
			continue
		}
		break
	}

	if _, err := r.client.LXC.Delete(
		data.Node.ValueString(),
		int(data.VMID.ValueInt64()),
		nil,
	); err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
		return
	}

	// check if deleted
	for true {
		time.Sleep(time.Second * 2)
		idAvailable, err := r.client.Cluster.IsVMIDAvailable(int(data.VMID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node lxc, got error: %s", err))
			return
		}
		if !idAvailable {
			continue
		}
		break
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "deleted an lxc resource")
}

// TODO: implement
func (r *LXCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
