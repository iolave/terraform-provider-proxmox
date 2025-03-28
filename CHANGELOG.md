# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]
### Added
- proxmox_node_lxc.nameserver property.

### Removed
- Windows arm support.

### Fixed
- proxmox_node_firewall_rule update method crash.
- proxmox_node_lxc issue present in the create/read methods that sometimes inverted networks while being retrieved, ending up in a weird tfstate.
- proxmox_node_lxc.networks[].computed_ip changes doesn't implies a destroy anymore.

### Changed
- proxmox_node_lxc now executes commands async.

## [v0.1.4]
### Fixed
- proxmox_node_firewall_rule read method crash when reading rule comments.

## [v0.1.3]
### Added
- proxmox_node_lxc.on_boot property.

## [v0.1.2]
### Fixed
- proxmox_node_lxc.features crash while converting it to proxmox client object.

### Added
- proxmox_node_lxc.unprivileged property.

### Removed
- template examples.
- docs.

## [v0.1.1]
### Fixed
- proxmox_node_lxc.features properties are now injected into the proxmox api request.

### Changed
- proxmox_node_lxc.features properies are now pointers.
- proxmox_node_lxc.features properies are now not computed.

## [v0.1.0]
### Added
- proxmox_version data source.
- proxmox_node_firewall_rules data source. 
- proxmox_node_firewall_rules resource. 
- proxmox_node_firewall_rule resource. 
- proxmox_node_firewall_rule resource. 
- proxmox_node_lxc resource.

[unreleased]: https://github.com/iolave/terraform-provider-proxmox/compare/v0.1.4...staging
[v0.1.4]: https://github.com/iolave/terraform-provider-proxmox/releases/tag/v0.1.4
[v0.1.3]: https://github.com/iolave/terraform-provider-proxmox/releases/tag/v0.1.3
[v0.1.2]: https://github.com/iolave/terraform-provider-proxmox/releases/tag/v0.1.2
[v0.1.1]: https://github.com/iolave/terraform-provider-proxmox/releases/tag/v0.1.1
[v0.1.0]: https://github.com/iolave/terraform-provider-proxmox/releases/tag/v0.1.0
