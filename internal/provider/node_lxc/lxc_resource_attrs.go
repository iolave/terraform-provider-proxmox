package nodelxc

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iolave/go-proxmox/pkg/pve"
)

func newLXCResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"node": schema.StringAttribute{
			Description: DESC_LXC_NODE,
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"os_template": schema.StringAttribute{
			Description: DESC_LXC_OSTEMPL,
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"id": schema.Int64Attribute{
			Description: DESC_LXC_ID,
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplaceIfConfigured(),
			},
		},
		//"arch": schema.StringAttribute{
		//	Description: DESC_LXC_ARCH,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     stringdefault.StaticString(DFLT_LXC_ARCH),
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"bwlimit": schema.Int64Attribute{
		//	Description: DESC_LXC_BWLIM,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"console_mode": schema.StringAttribute{
		//	Description: DESC_LXC_CMODE,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     stringdefault.StaticString(DFLT_LXC_CMODE),
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"console": schema.BoolAttribute{
		//	Description: DESC_LXC_CONSOLE,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     booldefault.StaticBool(DFLT_LXC_CONSOLE),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"cores": schema.Int64Attribute{
		//	Description: DESC_LXC_CORES,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"cpu_limit": schema.Int64Attribute{
		//	Description: DESC_LXC_CPULIM,
		//	Optional:    true,
		//	Computed:    true,
		//	Default:     int64default.StaticInt64(DFLT_LXC_CPULIM),
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"cpu_units": schema.Int64Attribute{
		//	Description: DESC_LXC_CPUUNI,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"debug": schema.BoolAttribute{
		//	Description: DESC_LXC_DEBUG,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     booldefault.StaticBool(DFLT_LXC_DEBUG),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"description": schema.StringAttribute{
		//	Description: DESC_LXC_DESC,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"force": schema.BoolAttribute{
		//	Description: DESC_LXC_FORCE,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"hookscript": schema.StringAttribute{
		//	Description: DESC_LXC_HOOK,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		"hostname": schema.StringAttribute{
			Description: DESC_LXC_HOSTNAME,
			Optional:    true,
			//Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		//"ignore_unpack_errors": schema.BoolAttribute{
		//	Description: DESC_LXC_IGNERR,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"lock": schema.StringAttribute{
		//	Description: DESC_LXC_LOCK,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"memory": schema.Int64Attribute{
		//	Description: DESC_LXC_MEM,
		//	Optional:    true,
		//	Computed:    true,
		//	Default:     int64default.StaticInt64(DFLT_LXC_MEM),
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"nameserver": schema.StringAttribute{
		//	Description: DESC_LXC_NS,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		"networks": schema.ListNestedAttribute{
			Description: DESC_LXC_NET,
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: newLXCNetResourceAttrs(),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"on_boot": schema.BoolAttribute{
			Description: DESC_LXC_ONBOOT,
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		//"os_type": schema.StringAttribute{
		//	Description: DESC_LXC_OSTYPE,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		"password": schema.StringAttribute{
			Description: DESC_LXC_PWD,
			Optional:    true,
			//Computed:    true,
			Sensitive: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		//"pool": schema.StringAttribute{
		//	Description: DESC_LXC_POOL,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"protection": schema.BoolAttribute{
		//	Description: DESC_LXC_PROTECTON,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     booldefault.StaticBool(DFLT_LXC_PROTECTON),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"restore": schema.BoolAttribute{
		//	Description: DESC_LXC_RESTORE,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"search_domain": schema.StringAttribute{
		//	Description: DESC_LXC_SDOMAIN,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		"ssh_public_keys": schema.ListAttribute{
			Description: DESC_LXC_SSH,
			Optional:    true,
			//Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		//"start": schema.BoolAttribute{
		//	Description: DESC_LXC_START,
		//	Optional:    true,
		//	Computed:    true,
		//	Default:     booldefault.StaticBool(DFLT_LXC_START),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"storage_id": schema.StringAttribute{
		//	Description: DESC_LXC_STORAGE,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     stringdefault.StaticString(DFLT_LXC_STORAGE),
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"swap_size": schema.Int64Attribute{
		//	Description: DESC_LXC_SWAP,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     int64default.StaticInt64(DFLT_LXC_SWAP),
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"tags": schema.ListAttribute{
		//	Description: DESC_LXC_TAGS,
		//	Optional:    true,
		//	//Computed:    true,
		//	ElementType: types.StringType,
		//	PlanModifiers: []planmodifier.List{
		//		listplanmodifier.RequiresReplace(),
		//	},
		//},
		//"template": schema.BoolAttribute{
		//	Description: DESC_LXC_TEMPLATE,
		//	Optional:    true,
		//	Computed:    true,
		//	Default:     booldefault.StaticBool(DFLT_LXC_TEMPLATE),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		//"timezone": schema.StringAttribute{
		//	Description: DESC_LXC_TZ,
		//	Optional:    true,
		//	//Computed:    true,
		//	PlanModifiers: []planmodifier.String{
		//		stringplanmodifier.RequiresReplace(),
		//	},
		//},
		//"available_tty": schema.Int64Attribute{
		//	Description: DESC_LXC_TTY,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     int64default.StaticInt64(DFLT_LXC_TTY),
		//	PlanModifiers: []planmodifier.Int64{
		//		int64planmodifier.RequiresReplace(),
		//	},
		//},
		//"unique_hw_addr": schema.BoolAttribute{
		//	Description: DESC_LXC_UNIQUE,
		//	Optional:    true,
		//	//Computed:    true,
		//	//Default:     booldefault.StaticBool(DFLT_LXC_UNIQUE),
		//	PlanModifiers: []planmodifier.Bool{
		//		boolplanmodifier.RequiresReplace(),
		//	},
		//},
		"unprivileged": schema.BoolAttribute{
			Description: DESC_LXC_UNPRIV,
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"cmds": schema.ListAttribute{
			Description: DESC_LXC_CMDS,
			ElementType: types.StringType,
			Optional:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"status": schema.StringAttribute{
			Description: DESC_LXC_STATUS,
			Computed:    true,
			Optional:    true,
			Default:     stringdefault.StaticString(string(pve.LXC_STATUS_STOPPED)),
		},
	}
}

// TODO: Add descriptions and default values
func newLXCFeaturesResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"force_rw_sys": schema.BoolAttribute{
			Optional: true,
		},
		"fuse": schema.BoolAttribute{
			Optional: true,
		},
		"key_ctl": schema.BoolAttribute{
			Optional: true,
		},
		"nesting": schema.BoolAttribute{
			Optional: true,
		},
	}
}

// TODO: Add descriptions and default values
func newLXCRootFSResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"volume": schema.StringAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"acl": schema.BoolAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"quota": schema.BoolAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"replicate": schema.BoolAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"read_only": schema.BoolAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"shared": schema.BoolAttribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"disk_size": schema.Int64Attribute{
			Optional: true,
			//Computed: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
	}
}

// TODO: Add descriptions and default values
func newLXCNetResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"bridge": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"firewall": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(DFLT_LXC_NET_FW),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"gateway": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"gateway6": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"hw_address": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ip": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"ip6": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"link_down": schema.BoolAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"mtu": schema.Int64Attribute{
			Optional: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"rate": schema.Int64Attribute{
			Optional: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"tag": schema.Int64Attribute{
			Optional: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"computed_ip": schema.StringAttribute{
			Description: DESC_LXC_IP,
			Computed:    true,
		},
	}
}
