# Contabo Terraform provider: the fixed and extended edition

Manage your whole Contabo account with Terraform or OpenTofu: servers, firewalls, DNS, reverse DNS, virtual IPs, object storage, users and roles. This provider starts from the [official Contabo provider](https://github.com/contabo/terraform-provider-contabo) and adds what the official provider is missing.

- **More of the Contabo API.** This provider has 8 resources and 12 data sources that the official provider does not have. With these resources and data sources, Terraform can run DNS for your domains, move a floating IP between servers, set up users with limited API access, and read your S3 keys.
- **Bugs fixed that can cost you servers or data.** The official provider cannot read a firewall that protects a server, loses your cloud-init data when it reinstalls a server, and can crash during an apply. This provider fixes these bugs and many more.
- **Safe with paid products.** A Contabo server costs money as soon as Terraform orders it. This provider saves a server in the Terraform state right after the order, so a failed apply does not lead to a second order. The provider can keep a server that you only want to stop managing, and the provider stops a plan early when the API cannot make a change.
- **Tested against the real API.** The new and the fixed resources have unit tests that run Terraform against a fake Contabo API. The provider was also tested against a real Contabo account, including a move of existing state from the official provider without a single planned change.
- **A drop-in replacement.** The resources use the same schema as the official provider. You change one line, run one command, and keep your existing Terraform state.

This provider is an unofficial distribution. The provider is not made by Contabo GmbH, and the provider is not connected to Contabo GmbH. Every change is also offered to Contabo as a pull request, and the file [PATCHES.md](PATCHES.md) shows the status of each change.

## Only in this provider

### Resources

| Resource | What you can do with the resource |
|---|---|
| `contabo_dns_zone` | Host the DNS zone of a domain on the Contabo name servers. |
| `contabo_dns_record` | Manage `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV` and `CAA` records in a DNS zone. |
| `contabo_ptr_record` | Set the reverse DNS name of an IPv4 or IPv6 address, which mail servers need. |
| `contabo_vip_assignment` | Assign a floating or additional IP address to a server, and move the address to another server. |
| `contabo_instance_reinstall` | Reinstall a server on purpose, for example after you attach a private network. |
| `contabo_instance_vnc` | Set a strong password for the VNC console of a server. |
| `contabo_role` | Create a role that allows only certain API calls, optionally only on resources with certain tags. |
| `contabo_user` | Create a user with such a role, for example a user that can only change DNS records. |
| `contabo_object_storage_object` | Upload a small file, such as a JSON document, into a bucket. The provider uses the S3 keys of its API user, so you need no S3 provider. |

### Data sources

| Data source | What you can read with the data source |
|---|---|
| `contabo_instances` | Search your servers by name, display name, product, region, status or tag. |
| `contabo_dns_zone` | Read a DNS zone and its name servers. |
| `contabo_ptr_record` | Read the reverse DNS name of an IP address. |
| `contabo_vip` and `contabo_vips` | Find your floating and additional IP addresses, for example the addresses that no server uses yet. |
| `contabo_object_storage_credentials` | Read the S3 access key and secret key, so that you never copy the keys by hand. |
| `contabo_image_stats` | Check how much custom image storage is free before you upload an image. |
| `contabo_instance_available_products` | List the products that a server can be upgraded to, with prices. |
| `contabo_instance_vnc` | Read the VNC address and port of a server. |
| `contabo_role`, `contabo_api_permissions` and `contabo_users` | Read roles, the API permissions that a role can grant, and the users of your account. |
| `contabo_kubernetes_oidc_documents` | Compute the discovery document and the key set of a Kubernetes service account issuer, so that you can publish them in a public bucket for workload identity. |

### New settings on existing resources

| Resource | New setting | What the setting does |
|---|---|---|
| `contabo_instance` | `cancel_on_destroy` | Keeps the server running when Terraform destroys the resource. |
| `contabo_instance` | `add_ons` | Orders add-ons with a new server, and adds private networking or backups to an existing server. |
| `contabo_instance` | `timeouts` | Limits how long Terraform waits for a server to install. |
| `contabo_firewall` | `display_name` on rules | Names firewall rules, as the Customer Control Panel does. |
| `contabo_firewall` data source | `name` | Finds a firewall by its name. |

## Bugs fixed

| Problem in the official provider | Status in this provider |
|---|---|
| The provider cannot read, import or create a firewall that protects a server, because the API sends the list of servers in an unexpected format ([#52](https://github.com/contabo/terraform-provider-contabo/issues/52), [#56](https://github.com/contabo/terraform-provider-contabo/issues/56)). | Fixed. |
| The provider crashes when a firewall rule has an empty `src_cidr` block ([#56](https://github.com/contabo/terraform-provider-contabo/issues/56)). | Fixed. |
| A reinstall after an image change loses the `user_data` value ([#49](https://github.com/contabo/terraform-provider-contabo/issues/49)). | Fixed. |
| A reinstall resets the `default_user` value to `admin` ([#45](https://github.com/contabo/terraform-provider-contabo/issues/45)). | Fixed. |
| A change of `product_id` does nothing, and the provider still reports success ([#56](https://github.com/contabo/terraform-provider-contabo/issues/56)). | The plan now stops with a clear error. |
| A server adopted with `existing_instance_id` cannot be destroyed ([#38](https://github.com/contabo/terraform-provider-contabo/issues/38)). | Fixed, and `cancel_on_destroy = false` keeps the server. |
| An apply can leave paid servers that Terraform does not know about ([#36](https://github.com/contabo/terraform-provider-contabo/issues/36)). | The provider saves the server id right after the order, and the provider warns about unexpected new servers. |
| An imported object storage shows a change in every plan ([#34](https://github.com/contabo/terraform-provider-contabo/issues/34)). | Fixed. |
| There is no way to list servers ([#54](https://github.com/contabo/terraform-provider-contabo/issues/54)). | The `contabo_instances` data source lists servers. |
| The provider cannot be built from the current API specification. | Fixed. |
| Errors from server updates and reinstalls are hidden, and the apply reports success. | Errors are reported. |
| A resource that someone deleted outside of Terraform breaks every plan. | Terraform plans to create the resource again. |
| Two provider configurations for two Contabo accounts share one login token. | Each account gets its own token, and long applies log in again when the token expires. |

The file [CHANGELOG.md](CHANGELOG.md) lists every change, and the file [PATCHES.md](PATCHES.md) links each change to its pull request for the official provider.

## Install

The Terraform Registry publishes the provider as [`hayone1/contabo`](https://registry.terraform.io/providers/hayone1/contabo/latest). The OpenTofu Registry publishes the provider under the same name.

## How to use the provider

Add the provider to the `required_providers` block of your Terraform configuration, and give the provider your API credentials:

```terraform
terraform {
  required_providers {
    contabo = {
      source  = "hayone1/contabo"
      version = "~> 0.2"
    }
  }
}

provider "contabo" {
  # You can also set the environment variables CNTB_OAUTH2_CLIENT_ID,
  # CNTB_OAUTH2_CLIENT_SECRET, CNTB_OAUTH2_USER and CNTB_OAUTH2_PASS
  # instead of writing the credentials here.
  oauth2_client_id     = var.contabo_client_id
  oauth2_client_secret = var.contabo_client_secret
  oauth2_user          = var.contabo_user
  oauth2_pass          = var.contabo_password
}
```

You can find your API credentials in the [Customer Control Panel](https://my.contabo.com/api/details) of your Contabo account.

**Warning about costs:** When Terraform creates a `contabo_instance` resource or a `contabo_object_storage` resource, Contabo orders the product and starts to bill you for the product. When Terraform destroys one of these resources, Contabo cancels the product at the end of the contract period. If Terraform must never cancel an instance, set `cancel_on_destroy = false` on that instance.

## How to move from the official provider

The resources of the distribution use the same schema as the resources of the official `contabo/contabo` provider. Because the schemas match, you can keep your existing Terraform state. Follow these steps:

1. In the `required_providers` block, change `source = "contabo/contabo"` to `source = "hayone1/contabo"`.
2. Run `terraform init -upgrade`.
3. Run `terraform state replace-provider registry.terraform.io/contabo/contabo registry.terraform.io/hayone1/contabo`. If you use OpenTofu, write `registry.opentofu.org` instead of `registry.terraform.io` in both addresses.
4. Run `terraform plan` and read the plan carefully before you apply the plan. The file [CHANGELOG.md](CHANGELOG.md) explains which behavior changed.

## How to develop the provider

To work on the provider, you need Go (the version that the file `go.mod` names) and Terraform or OpenTofu for the tests. You need Docker only when you want to generate the API client again.

These `make` targets help with the daily work:

```sh
make build          # Build the provider binary.
make test           # Run the unit tests against an in-memory fake of the Contabo API.
make lint           # Run go vet and golangci-lint.
make docs           # Generate the docs/ folder from the schema, templates/ and examples/.
```

The Go API client in the `openapi/` folder is generated from a copy of the Contabo API specification. The copy is stored in the file `spec/api-v1.yaml`. These two targets update the copy and the client:

```sh
make update-spec           # Download the current API specification into spec/api-v1.yaml.
make generate-api-clients  # Generate the openapi/ folder from spec/api-v1.yaml. This target needs Docker.
```

To test a local build with Terraform, add a development override to the file `~/.terraformrc`. The override tells Terraform to use your local binary instead of the binary from the registry:

```terraform
provider_installation {
  dev_overrides {
    "hayone1/contabo" = "/path/to/the/directory/with/the/binary"
  }
  direct {}
}
```

### Acceptance tests

The acceptance tests use the real Contabo API. The acceptance tests order real products, so the acceptance tests cost money. Run the acceptance tests with this command:

```sh
make test-acc
```

The acceptance tests need the environment variables `CNTB_API`, `CNTB_OAUTH2_TOKEN_URL`, `CNTB_OAUTH2_CLIENT_ID`, `CNTB_OAUTH2_CLIENT_SECRET`, `CNTB_OAUTH2_USER` and `CNTB_OAUTH2_PASS`. The DNS tests and the reverse DNS tests only run when you also set the following environment variables:

- `CNTB_TEST_DNS_ZONE` is a domain that the Contabo name servers do not host yet. The test creates a DNS zone for this domain and deletes the DNS zone at the end.
- `CNTB_TEST_PTR_IPV6` is an unused IPv6 address of one of your instances, for example `<instance /64 prefix>::100`. The test creates a PTR record for this address and deletes the PTR record at the end.
- `CNTB_TEST_PTR_IPV4` is an IPv4 address of one of your instances. The test changes the PTR record of this address and restores the old PTR record at the end.

### Releases

The file [RELEASING.md](RELEASING.md) explains how to create the signing key, how to publish a release, and how to publish the provider on the Terraform Registry and the OpenTofu Registry.

### The two repositories

The work happens in two repositories:

- [`hayone1/terraform-provider-contabo`](https://github.com/hayone1/terraform-provider-contabo) is this repository. This repository holds the distribution. The distribution follows the official repository with merges and carries the extra changes.
- [`hayone1/terraform-provider-contabo-contrib`](https://github.com/hayone1/terraform-provider-contabo-contrib) is a fork of the official repository. Pull requests to the official provider come from this fork.

## License

The provider uses the MIT license. The file [LICENSE](LICENSE) has the full text. Contabo GmbH holds the copyright for the code of the official provider.
