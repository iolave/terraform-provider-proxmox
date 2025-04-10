package lxc

const DESC_RSRC_LXC_EXEC = `lxc_exec execute commands inside an lxc.

- If a command fails, the execution will stop and the resource will not be created.`
const MD_RSRC_LXC_EXEC = `**lxc_exec** execute commands inside an lxc.

- If a command fails, the execution will stop and the resource will not be created.`

const DESC_LXC = "LXC is a well-known Linux " +
	"container runtime that consists of tools, " +
	"templates, and library and language bindings. " +
	"It's pretty low level, very flexible and covers " +
	"just about every containment feature supported " +
	"by the upstream kernel."
const DESC_LXC_NODE = "The cluster node name."
const DESC_LXC_OSTEMPL = "The OS template or backup file."
const DESC_LXC_ID = "The (unique) ID of the VM."
const DESC_LXC_ARCH = "OS architecture type. \n" +
	"Values: amd64 | i386 | arm64 | armhf | riscv32 | riscv64"
const DFLT_LXC_ARCH = "amd64"
const DESC_LXC_BWLIM = "Override I/O bandwidth limit (in KiB/s)."
const DESC_LXC_CMODE = "Console mode. By default, the console " +
	"command tries to open a connection to one of the " +
	"available tty devices. By setting cmode to " +
	"'console' it tries to attach to /dev/console " +
	"instead. If you set cmode to 'shell', it simply " +
	"invokes a shell inside the container (no login).\n" +
	"Values: shell | console | tty"
const DFLT_LXC_CMODE = "tty"
const DESC_LXC_CONSOLE = "Attach a console device (/dev/console) " +
	"to the container."
const DFLT_LXC_CONSOLE = true
const DESC_LXC_CORES = "The number of cores assigned to the container. " +
	"A container can use all available cores by default."
const DESC_LXC_CPULIM = "Limit of CPU usage.\n" +
	"NOTE: If the computer has 2 CPUs, it has a total of " +
	"'2' CPU time. Value '0' indicates no CPU limit."
const DFLT_LXC_CPULIM = 0
const DESC_LXC_CPUUNI = "CPU weight for a container. Argument " +
	"is used in the kernel fair scheduler. The larger " +
	"the number is, the more CPU time this container " +
	"gets. Number is relative to the weights of all the " +
	"other running guests."
const DESC_LXC_DEBUG = "Try to be more verbose. For now this only " +
	"enables debug log-level on start."
const DFLT_LXC_DEBUG = false
const DESC_LXC_DESC = "Description for the Container. Shown " +
	"in the web-interface CT's summary. This is saved " +
	"as comment inside the configuration file."
const DESC_LXC_FEATS = "Allow containers access to advanced features."

// TODO: Add features descriptions
const DESC_LXC_FORCE = "Allow to overwrite existing container."
const DESC_LXC_HOOK = "Script that will be executed during various " +
	"steps in the containers lifetime."
const DESC_LXC_HOSTNAME = "Set a host name for the container."
const DESC_LXC_IGNERR = "Ignore errors when extracting the template."
const DESC_LXC_LOCK = "Lock/unlock the container.\n" +
	"Values: backup | create | destroyed | disk | fstrim | " +
	"migrate | mounted | rollback | snapshot | snapshot-delete"
const DESC_LXC_MEM = "Amount of RAM for the container in MB."
const DFLT_LXC_MEM = 512

// TODO: Add mp[n] descriptions
const DESC_LXC_NS = "Sets DNS server IP address for a container. " +
	"Create will automatically use the setting from the host " +
	"if you neither set searchdomain nor nameserver."
const DESC_LXC_NET = "Specifies network interface for the container."
const DFLT_LXC_NET_FW = false
const DESC_LXC_ONBOOT = "Specifies whether a container will be " +
	"started during system bootup."
const DFLT_LXC_ONBOOT = false
const DESC_LXC_OSTYPE = "OS type. This is used to setup " +
	"configuration inside the container, and corresponds to " +
	"lxc setup scripts in " +
	"/usr/share/lxc/config/<ostype>.common.conf. Value " +
	"'unmanaged' can be used to skip and OS specific setup.\n" +
	"Values: debian | devuan | ubuntu | centos | fedora | " +
	"opensuse | archlinux | alpine | gentoo | nixos | unmanaged"
const DESC_LXC_PWD = "Sets root password inside container."
const DESC_LXC_POOL = "Add the VM to the specified pool."
const DESC_LXC_PROTECTON = "Sets the protection flag of the container." +
	"This will prevent the CT or CT's disk remove/update " +
	"operation."
const DFLT_LXC_PROTECTON = false
const DESC_LXC_RESTORE = "Mark this as restore task."
const DESC_LXC_ROOTFS = "Use volume as container root."

// TODO: Add root_fs.xyz descriptions
const DESC_LXC_SDOMAIN = "Sets DNS search domains for a container. " +
	"Create will automatically use the setting from the " +
	"host if you neither set searchdomain nor nameserver."
const DESC_LXC_SSH = "Setup public SSH keys (OpenSSH format)."
const DESC_LXC_START = "Start the CT after its creation finished " +
	"successfully."
const DFLT_LXC_START = false
const DESC_LXC_STARTUP = "Startup and shutdown behavior. Order is a " +
	"non-negative number defining the general startup order. " +
	"Shutdown in done with reverse ordering. Additionally " +
	"you can set the 'up' or 'down' delay in seconds, which " +
	"specifies a delay to wait before the next VM is started " +
	"or stopped."
const DESC_LXC_STORAGE = "Default Storage."
const DFLT_LXC_STORAGE = "local"
const DESC_LXC_SWAP = "Amount of SWAP for the container in MB."
const DFLT_LXC_SWAP = 512
const DESC_LXC_TAGS = "Tags of the Container. This is only " +
	"meta information."
const DESC_LXC_TEMPLATE = "Enable/disable Template."
const DFLT_LXC_TEMPLATE = false
const DESC_LXC_TZ = "Time zone to use in the containerw If " +
	"option isn't set, then nothing will be done. Can be " +
	"set to 'host' to match the host time zone, or an " +
	"arbitrary time zone option from " +
	"/usr/share/zoneinfo/zone.tab"
const DESC_LXC_TTY = "Specify the number of tty available to the " +
	"container."
const DFLT_LXC_TTY = 2
const DESC_LXC_UNIQUE = "Assign a unique random ethernet address."
const DFLT_LXC_UNIQUE = true
const DESC_LXC_UNPRIV = "Makes the container run as unprivileged user." +
	"(Should not be modified manually.)"
const DFLT_LXC_UNPRIV = false
const DESC_LXC_STATUS = "LXC Container status.\n" +
	"Values: stopped | running"
const DESC_LXC_IP = "LXC assigned ip."
const DESC_LXC_CMDS = "List of commands to be executed after lxc " +
	"creation using bash. If any command fail, the creation " +
	"will also fail."
