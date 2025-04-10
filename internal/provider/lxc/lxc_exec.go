package lxc

import (
	"context"
	"fmt"
	nodelxc "terraform-provider-proxmox/internal/provider/node_lxc"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iolave/go-proxmox/pkg/pve"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LXCExecResource{}
var _ resource.ResourceWithImportState = &LXCExecResource{}

func NewLXCExecResource() resource.Resource {
	return &LXCExecResource{}
}

// LXCExecResource defines the resource implementation.
type LXCExecResource struct {
	client *pve.PVE
}

// LXCExecResourceModel describes the resource data model.
type LXCExecResourceModel struct {
	VMID types.Int64    `tfsdk:"id"`
	CMDs []types.String `tfsdk:"cmds"`
}

func (r *LXCExecResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	name := "lxc_exec"
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, name)
}

func (r *LXCExecResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: MD_RSRC_LXC_EXEC,
		Description:         DESC_RSRC_LXC_EXEC,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: nodelxc.DESC_LXC_ID,
				Required:    true,
			},
			"cmds": schema.ListAttribute{
				Description: nodelxc.DESC_LXC_CMDS,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *LXCExecResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LXCExecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LXCExecResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmid := int(data.VMID.ValueInt64())

	if err := nodelxc.RunLXCCommands(ctx, r.client, vmid, data.CMDs); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to run commands inside lxc , got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LXCExecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *LXCExecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *LXCExecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *LXCExecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
