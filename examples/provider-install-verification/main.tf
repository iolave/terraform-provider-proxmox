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

data "proxmox_version" "example" {
}

output "example_version" {
  value = data.proxmox_version.example
}
