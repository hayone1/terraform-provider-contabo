# terraform-provider-contabo (hayone1 distribution)

Unofficial distribution of the Contabo Terraform provider. Not made by or connected to Contabo GmbH.

This repository is the [upstream provider](https://github.com/contabo/terraform-provider-contabo) plus fixes and features that are not merged upstream yet. [PATCHES.md](PATCHES.md) lists every carried patch with its upstream pull request or issue. When upstream merges a patch, this distribution drops its own copy. [CHANGELOG.md](CHANGELOG.md) lists the changes per release.

- Terraform Registry: [`hayone1/contabo`](https://registry.terraform.io/providers/hayone1/contabo/latest)
- OpenTofu Registry: `hayone1/contabo`

## Usage

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
  # Or set CNTB_OAUTH2_CLIENT_ID, CNTB_OAUTH2_CLIENT_SECRET,
  # CNTB_OAUTH2_USER and CNTB_OAUTH2_PASS.
  oauth2_client_id     = var.contabo_client_id
  oauth2_client_secret = var.contabo_client_secret
  oauth2_user          = var.contabo_user
  oauth2_pass          = var.contabo_password
}
```

The credentials are in the [Customer Control Panel](https://my.contabo.com/api/details).

**Costs:** creating a `contabo_instance` or `contabo_object_storage` orders a product and starts billing. Destroying it cancels the product at the end of its contract period. Set `cancel_on_destroy = false` on instances that must not be cancelled by Terraform.

## What this distribution adds

Highlights; see [PATCHES.md](PATCHES.md) for the full list.

- Fixes: the API client builds again, firewalls with instances can be read and created, reinstalls keep `user_data` and `default_user`, errors from instance updates are reported, resources deleted outside Terraform no longer break plans, object storage imports converge.
- Instances: `cancel_on_destroy`, safe adoption with `existing_instance_id`, add-on upgrades (private networking, backup), plan-time errors for changes the API cannot make, timeouts.
- New resources: DNS zones and records, reverse DNS (PTR), `contabo_instance_reinstall`, `contabo_vip_assignment`, `contabo_role`, `contabo_user`.
- New data sources: `contabo_instances`, `contabo_vip(s)`, `contabo_object_storage_credentials`, `contabo_image_stats`, `contabo_instance_available_products`, `contabo_instance_vnc`, `contabo_role`, `contabo_api_permissions`, `contabo_users`, `contabo_dns_zone`, `contabo_ptr_record`.

## Migrating from `contabo/contabo`

Resource schemas are compatible with the official provider, so existing state can be moved:

1. Change `source = "contabo/contabo"` to `source = "hayone1/contabo"`.
2. Run `terraform init -upgrade`.
3. Run `terraform state replace-provider registry.terraform.io/contabo/contabo registry.terraform.io/hayone1/contabo` (with OpenTofu, use `registry.opentofu.org` in both addresses).
4. Run `terraform plan` and review it. [CHANGELOG.md](CHANGELOG.md) lists the behavior changes.

## Development

Requirements: Go (version in `go.mod`), Terraform or OpenTofu for the tests, Docker only to regenerate the API client.

```sh
make build          # build the provider binary
make test           # unit tests against an in-memory fake of the Contabo API
make lint           # go vet and golangci-lint
make docs           # regenerate docs/ from the schema, templates/ and examples/
```

The Go API client in `openapi/` is generated from the committed spec snapshot `spec/api-v1.yaml`:

```sh
make update-spec           # download the current spec
make generate-api-clients  # regenerate openapi/ (needs Docker)
```

To use a local build, add a development override to `~/.terraformrc`:

```terraform
provider_installation {
  dev_overrides {
    "hayone1/contabo" = "/path/to/the/directory/with/the/binary"
  }
  direct {}
}
```

### Acceptance tests

Acceptance tests use the real API and order real resources, which costs money:

```sh
make test-acc
```

They need `CNTB_API`, `CNTB_OAUTH2_TOKEN_URL`, `CNTB_OAUTH2_CLIENT_ID`, `CNTB_OAUTH2_CLIENT_SECRET`, `CNTB_OAUTH2_USER` and `CNTB_OAUTH2_PASS`. The DNS and reverse DNS acceptance tests are skipped unless these are set:

- `CNTB_TEST_DNS_ZONE`: a domain that is not yet hosted on the Contabo name servers. The zone is created and deleted by the test.
- `CNTB_TEST_PTR_IPV6`: an unused IPv6 address of one of your instances, for example `<instance /64 prefix>::100`. A PTR record is created and deleted for it.
- `CNTB_TEST_PTR_IPV4`: an IPv4 address of one of your instances. Its PTR record is changed during the test and restored afterwards.

### Repositories

- [`hayone1/terraform-provider-contabo`](https://github.com/hayone1/terraform-provider-contabo) (this repository): the distribution. It tracks `contabo/terraform-provider-contabo` with merges and carries the patches.
- [`hayone1/terraform-provider-contabo-contrib`](https://github.com/hayone1/terraform-provider-contabo-contrib): the fork that upstream pull requests come from.

## License

MIT, see [LICENSE](LICENSE). The upstream code is copyright Contabo GmbH.
