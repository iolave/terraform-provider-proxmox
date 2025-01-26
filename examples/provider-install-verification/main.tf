terraform {
  required_providers {
    proxmox = {
      source = "hashicorp.com/edu/proxmox"
    }
  }
}

provider "proxmox" {
  // host                 = "localhost"
  // port                 = 8006
  // user                 = "root@pam"
  // token                = ""
  // token_name           = ""
  // cf_client_id         = ""
  // cf_client_secret     = ""
  insecure_skip_verify = true
}

// data "proxmox_version" "example" {
// }
// 
// output "example_version" {
//   value = data.proxmox_version.example
// }
// 
// data "proxmox_node_firewall_rules" "node_fw_rules" {
//   node = "pve-prd-1"
// }
// 
// output "node_fw_rules" {
//   value = data.proxmox_node_firewall_rules.node_fw_rules
// }

resource "proxmox_node_firewall_rules" "example" {
  node = "pve-prd-1"
  rules = [
    {
      action      = "ACCEPT"
      type        = "in"
      comment     = "TF_CREATED"
      macro       = "SSH"
      destination = "dc/pve-prd-1-local"
    },
    {
      action      = "ACCEPT"
      type        = "in"
      comment     = "TF_CREATED"
      macro       = "SSH"
      destination = "dc/pve-prd-1-local"
    }
  ]
}
