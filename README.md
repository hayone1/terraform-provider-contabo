# terraform-provider-contabo (hayone1 distribution)

This repository is an unofficial distribution of the Contabo Terraform provider. This distribution is not made by Contabo GmbH, and it is not connected to Contabo GmbH.

The distribution starts from the [official provider](https://github.com/contabo/terraform-provider-contabo) that Contabo publishes. The distribution then adds fixes and new features that Contabo has not merged into the official provider yet. The file [PATCHES.md](PATCHES.md) lists every one of these changes, together with the pull request or issue in the official repository. When Contabo merges one of these changes into the official provider, the distribution stops carrying its own copy of that change. The file [CHANGELOG.md](CHANGELOG.md) lists the changes in each release.

You can install the distribution from two registries:

- The Terraform Registry publishes the distribution as [`hayone1/contabo`](https://registry.terraform.io/providers/hayone1/contabo/latest).
- The OpenTofu Registry publishes the distribution as `hayone1/contabo` too.

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

## What the distribution adds

This list shows the most important changes. The file [PATCHES.md](PATCHES.md) has the complete list.

- **Fixes:** The API client builds again. The provider can read and create firewalls that protect instances. A reinstall keeps the `user_data` value and the `default_user` value. The provider reports errors when an instance update fails. A resource that someone deleted outside of Terraform no longer breaks the plan. An imported object storage no longer shows changes in every plan.
- **Instances:** The new `cancel_on_destroy` setting decides whether destroying an instance cancels the instance. The `existing_instance_id` setting adopts an existing instance without reinstalling the instance. The provider can add private networking and backups to an existing instance. The plan fails early when you ask for a change that the API cannot make. Long operations have time limits.
- **New resources:** DNS zones, DNS records, reverse DNS (PTR) records, `contabo_instance_reinstall`, `contabo_vip_assignment`, `contabo_role` and `contabo_user`.
- **New data sources:** `contabo_instances`, `contabo_vip`, `contabo_vips`, `contabo_object_storage_credentials`, `contabo_image_stats`, `contabo_instance_available_products`, `contabo_instance_vnc`, `contabo_role`, `contabo_api_permissions`, `contabo_users`, `contabo_dns_zone` and `contabo_ptr_record`.

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

### The two repositories

The work happens in two repositories:

- [`hayone1/terraform-provider-contabo`](https://github.com/hayone1/terraform-provider-contabo) is this repository. This repository holds the distribution. The distribution follows the official repository with merges and carries the extra changes.
- [`hayone1/terraform-provider-contabo-contrib`](https://github.com/hayone1/terraform-provider-contabo-contrib) is a fork of the official repository. Pull requests to the official provider come from this fork.

## License

The provider uses the MIT license. The file [LICENSE](LICENSE) has the full text. Contabo GmbH holds the copyright for the code of the official provider.
