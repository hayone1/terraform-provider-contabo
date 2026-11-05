# Carried patches

This distribution is [`contabo/terraform-provider-contabo`](https://github.com/contabo/terraform-provider-contabo) plus the patches below. When upstream merges a patch, the next sync resolves conflicts in favor of upstream and the patch is marked `merged upstream`.

Status values: `carried` (in this distribution, not upstream), `merged upstream`, `dropped`.

Upstream base: `v0.1.44`.

| # | Patch | Source in this repository | Upstream PR or issue | Status | Since |
|---|---|---|---|---|---|
| 1 | API client generation with operations that have several tags | merge of upstream PR #57 (`4b7d067`) | [PR #57](https://github.com/contabo/terraform-provider-contabo/pull/57) | carried | v0.2.0 |
| 1a | Generate the client from a committed spec snapshot (`spec/api-v1.yaml`) and commit the generated `openapi/` | `d27022b` | — (distribution only) | carried | v0.2.0 |
| 2 | DNS zone, DNS record and reverse DNS (PTR) resources and data sources | merge of `contrib/feature/dns-and-reverse-dns` (`b81e315`) | [PR #59](https://github.com/contabo/terraform-provider-contabo/pull/59) | carried | v0.2.0 |
| 3 | Keep `default_user` on reinstall | merge of upstream PR #47 (`95cbad8`), completed by `f7ccce5` | [PR #47](https://github.com/contabo/terraform-provider-contabo/pull/47), [issue #45](https://github.com/contabo/terraform-provider-contabo/issues/45) | carried | v0.2.0 |
| 4 | Import for `contabo_object_storage_bucket` | merge of upstream PR #58 (`a9e4414`) | [PR #58](https://github.com/contabo/terraform-provider-contabo/pull/58) | carried | v0.2.0 |
| 5 | Firewall: decode the JSON-encoded `instances` string, fix the `src_cidr` panic and `instances_status`, assign only changed instances, hide the implicit drop-all rule, data source lookup by name | `394b3eb`, `dc3eb26` | [issue #52](https://github.com/contabo/terraform-provider-contabo/issues/52), [issue #56](https://github.com/contabo/terraform-provider-contabo/issues/56) | carried | v0.2.0 |
| 6 | Report errors from instance updates and reinstalls | `f7ccce5` | — | carried | v0.2.0 |
| 7 | Instances adopted with `existing_instance_id` can be destroyed; adoption does not reinstall | `f7ccce5` | [issue #38](https://github.com/contabo/terraform-provider-contabo/issues/38) | carried | v0.2.0 |
| 8 | `cancel_on_destroy` for instances | `f7ccce5`, `b1e1079` | — | carried | v0.2.0 |
| 9 | `product_id` and `region` changes fail at plan time | `f7ccce5` | [issue #56](https://github.com/contabo/terraform-provider-contabo/issues/56) | carried | v0.2.0 |
| 10 | Reinstalls send every reinstall attribute, so `user_data` is not lost | `f7ccce5` | [issue #49](https://github.com/contabo/terraform-provider-contabo/issues/49) | carried | v0.2.0 |
| 11 | Duplicate orders: save the id right after the order, bounded waits, warn about other new instances with the same name | `f7ccce5` | [issue #36](https://github.com/contabo/terraform-provider-contabo/issues/36) | carried (mitigation; the provider sends one order) | v0.2.0 |
| 12 | `add_ons` on create; private networking and backup on existing instances through the upgrade endpoint | `f7ccce5` | — | carried | v0.2.0 |
| 13 | `contabo_instance_reinstall` resource | `6c294bc` | — | carried | v0.2.0 |
| 14 | VIPs: `contabo_vip_assignment`, `contabo_vip`, `contabo_vips` | `6c294bc` | — | carried | v0.2.0 |
| 15 | `contabo_object_storage_credentials` data source | `6c294bc` | — | carried | v0.2.0 |
| 16 | `contabo_image_stats` data source | `6c294bc` | — | carried | v0.2.0 |
| 17 | `contabo_instances` data source | `f7ccce5` | [issue #54](https://github.com/contabo/terraform-provider-contabo/issues/54) | carried | v0.2.0 |
| 18 | `contabo_instance_available_products` data source | `6c294bc` | — | carried | v0.2.0 |
| 19 | `contabo_role`, `contabo_user`, `contabo_role` and `contabo_api_permissions` and `contabo_users` data sources | `6c294bc` | — | carried | v0.2.0 |
| 20 | Import for `contabo_object_storage` (region name and slug are equivalent) | `62d685a` | [issue #34](https://github.com/contabo/terraform-provider-contabo/issues/34) | carried | v0.2.0 |
| 21 | VNC: `contabo_instance_vnc` data source (read only) | `6c294bc` | [issue #51](https://github.com/contabo/terraform-provider-contabo/issues/51) | carried (partial: disabling VNC is not in the public API) | v0.2.0 |
| 22 | Per-account token cache, token refresh and re-login, detailed API errors, user id from the token | `709a3f6` | — | carried | v0.2.0 |
| 23 | Resources deleted outside Terraform are removed from state; image download wait with timeout | `62d685a` | — | carried | v0.2.0 |
| 24 | terraform-plugin-sdk v2.40, release pipeline for `hayone1/contabo`, CI, unit tests against a fake API | `a9cf95f`, `71fbcc9` | — (distribution only) | carried | v0.2.0 |

## Syncing with upstream

```sh
git fetch upstream
git merge upstream/main     # merge, never rebase: published tags must stay reachable
```

When upstream merged one of the patches (often squashed), the merge conflicts on that patch's files. Resolve in favor of upstream, set the status above to `merged upstream`, and delete the branch in the contrib fork.

## Adding a patch

- Meant for upstream: in the contrib clone, `git switch -c feature/<name> upstream/main`, commit, push, open the pull request. Then here: `git fetch contrib && git merge --no-ff contrib/feature/<name>`.
- Distribution only, or declined upstream: commit on a `patch/<name>` branch here and merge it to `main`.
- Third-party upstream pull requests: `git fetch upstream pull/<n>/head:upstream-pr-<n>`, then merge.
