# Terraform Provider Proxmox 

## Requirements
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22.7

## Building The Provider
1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies
This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider
### Environment variables
| Env | Default value |Description|
|-----|:---------------:|-----------|
|PROXMOX_HOST|127.0.0.1|Proxmox host|
|PROXMOX_PORT|8006|Proxmox port|
|PROXMOX_USER||Proxmox user (ie. `root@pam`)|
|PROXMOX_TOKEN||Proxmox user generated token|
|PROXMOX_TOKEN_NAME||Proxmox user generated token name|
|CF_CLIENT_ID||Cloudflare client id (when proxmox is secured by cloudflare)|
|CF_CLIENT_SECRET||Cloudflare client secret (when proxmox is secured by cloudflare)|


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
