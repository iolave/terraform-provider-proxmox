package nodefirewall

// descriptions for rules
const (
	DESC_RULE = "Firewall rule at a node level.\n" +
		"In order for a firewall rule to take effect " +
		"the pve node firewall must be enabled " +
		"otherwise the rule will be created but " +
		"it will not take effect."
	DESC_RULES     = "Array of rules.\n Rule: " + DESC_RULE
	DESC_RULE_NODE = "The cluster node name."
	DESC_RULE_ID   = "go-proxmox generated id that lives " +
		"within the rule comment field in proxmox."
	DESC_RULE_ACTION = "Rule action ('ACCEPT', 'DROP', " +
		"'REJECT') or security group name.\n" +
		`Format: [A-Za-z][A-Za-z0-9\-\_]+`
	DESC_RULE_COMMENT = "Descriptive comment.\n" +
		"Note: an id is prefixed to the comment field " +
		"within proxmox by the api client."
	DESC_RULE_DEST = "Restrict packet destination address. " +
		"This can refer to a single IP address, an IP " +
		"set ('+ipsetname') or an IP alias definition. " +
		"You can also specify an address range like " +
		"'20.34.101.207-201.3.9.99', or a list of IP " +
		"addresses and networks (entries are separated " +
		"by comma). Please do not mix IPv4 and IPv6 " +
		"addresses inside such lists."
	DESC_RULE_DPORT = "Restrict TCP/UDP destination port. " +
		"You can use service names or simple " +
		"numbers (0-65535), as defined in '/etc/services'. " +
		"Port ranges can be specified " +
		`with '\d+:\d+', for example '80:85', and you ` +
		"can use comma separated list to match several " +
		"ports or ranges."
	DESC_RULE_ENABLE = "Flag to enable/disable a rule."
	DESC_RULE_ICMP   = "Specify icmp-type. Only valid if proto " +
		"equals 'icmp' or 'icmpv6'/'ipv6-icmp'."
	DESC_RULE_IFACE = "Network interface name. You have to use " +
		"network configuration key names for " +
		`VMs and containers ('net\d+'). Host related rules ` +
		`can use arbitrary strings.`
	DESC_RULE_LOG = "Log level for firewall rule.\n" +
		"Values: emerg | alert | crit | err | warning " +
		"| notice | info | debug | nolog"
	DESC_RULE_MACRO = "Use predefined standard macro."
	DESC_RULE_POS   = "Update rule at position <pos>." +
		"Note: for some reason this doesn't work, might " +
		"be an api client issue."
	DESC_RULE_PROTO = "IP protocol. You can use protocol names " +
		"('tcp'/'udp') or simple numbers, as defined in " +
		"'/etc/protocols'."
	DESC_RULE_SOURCE = "Restrict packet source address. " +
		"This can refer to a single IP address, an IP " +
		"set ('+ipsetname') or an IP alias definition. " +
		"You can also specify an address range like " +
		"'20.34.101.207-201.3.9.99', or a list of IP " +
		"addresses and networks (entries are separated " +
		"by comma). Please do not mix IPv4 and IPv6 " +
		"addresses inside such lists."
	DESC_RULE_SPORT = "Restrict TCP/UDP source port. You can " +
		"use service names or simple numbers (0-65535), " +
		"as defined in '/etc/services'. Port ranges can " +
		`be specified with '\d+:\d+', for example '80:85', ` +
		"and you can use comma separated list to match " +
		"several ports or ranges."
	DESC_RULE_TYPE = "Rule type.\n" +
		"Values: in | out | forward | group"
)
