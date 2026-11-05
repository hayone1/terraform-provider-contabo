# Carried patches

This distribution is the official provider [`contabo/terraform-provider-contabo`](https://github.com/contabo/terraform-provider-contabo) plus the patches in the table below. A patch is a change that the official provider does not have yet. When Contabo merges a patch into the official provider, the next update from the official provider replaces the copy of the patch in this repository, and the table marks the patch as `merged upstream`.

The status column uses three values:

- `carried` means that this distribution contains the patch and the official provider does not.
- `merged upstream` means that the official provider now contains the patch.
- `dropped` means that this distribution no longer contains the patch.

The distribution is based on version `v0.1.44` of the official provider.

The pull requests #60 to #64 in the official repository come from the contrib fork. Each of these pull requests builds on the pull requests before it, starting with #59. The pull requests contain the same changes as the commits in this table, adapted to the older terraform-plugin-sdk version of the official provider.

| # | What the patch does | Commit in this repository | Pull request or issue in the official repository | Status | Since |
|---|---|---|---|---|---|
| 1 | The patch makes the API client build from the current API specification, which has operations with several tags. | Merge of pull request #57 (`4b7d067`) | [PR #57](https://github.com/contabo/terraform-provider-contabo/pull/57) | carried | v0.2.0 |
| 1a | The patch generates the API client from a stored copy of the API specification (`spec/api-v1.yaml`) and stores the generated client in `openapi/`. | `d27022b` | None. This patch only exists in the distribution. | carried | v0.2.0 |
| 2 | The patch adds resources and data sources for DNS zones, DNS records and reverse DNS (PTR) records. | Merge of `contrib/feature/dns-and-reverse-dns` (`b81e315`) | [PR #59](https://github.com/contabo/terraform-provider-contabo/pull/59) | carried | v0.2.0 |
| 3 | The patch keeps the `default_user` value when an instance is reinstalled. | Merge of pull request #47 (`95cbad8`), completed by `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [PR #47](https://github.com/contabo/terraform-provider-contabo/pull/47), [issue #45](https://github.com/contabo/terraform-provider-contabo/issues/45) | carried | v0.2.0 |
| 4 | The patch lets you import a `contabo_object_storage_bucket` resource. | Merge of pull request #58 (`a9e4414`) | [PR #58](https://github.com/contabo/terraform-provider-contabo/pull/58) | carried | v0.2.0 |
| 5 | The patch fixes firewalls. The provider reads the instance list that the API sends as JSON text, no longer crashes on an empty `src_cidr` block, sets `instances_status` correctly, changes only the instances that changed, hides the drop-all rule that the API adds, and lets the data source find a firewall by name. | `394b3eb`, `dc3eb26` | [PR #61](https://github.com/contabo/terraform-provider-contabo/pull/61), [Issue #52](https://github.com/contabo/terraform-provider-contabo/issues/52), [issue #56](https://github.com/contabo/terraform-provider-contabo/issues/56) | carried | v0.2.0 |
| 6 | The patch reports errors when an instance update or an instance reinstall fails. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62) | carried | v0.2.0 |
| 7 | The patch lets you destroy an instance that you adopted with `existing_instance_id`. The adoption no longer reinstalls the instance. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [Issue #38](https://github.com/contabo/terraform-provider-contabo/issues/38) | carried | v0.2.0 |
| 8 | The patch adds the `cancel_on_destroy` setting to instances. | `f7ccce5`, `b1e1079` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62) | carried | v0.2.0 |
| 9 | The patch makes the plan fail when you change `product_id` or `region` of an existing instance. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [Issue #56](https://github.com/contabo/terraform-provider-contabo/issues/56) | carried | v0.2.0 |
| 10 | The patch sends every reinstall setting during a reinstall, so the instance keeps its `user_data` value. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [Issue #49](https://github.com/contabo/terraform-provider-contabo/issues/49) | carried | v0.2.0 |
| 11 | The patch lowers the risk of duplicate orders. The provider saves the instance id right after the order, waits with a time limit, and warns about other new instances with the same name. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [Issue #36](https://github.com/contabo/terraform-provider-contabo/issues/36) | carried. The provider already sent only one order, so this patch is a safety measure. | v0.2.0 |
| 12 | The patch orders the `add_ons` of an instance when the instance is created, and adds private networking and backups to existing instances through the upgrade API. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62) | carried | v0.2.0 |
| 13 | The patch adds the `contabo_instance_reinstall` resource. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 14 | The patch adds the `contabo_vip_assignment` resource and the `contabo_vip` and `contabo_vips` data sources for virtual IP addresses. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 15 | The patch adds the `contabo_object_storage_credentials` data source. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 16 | The patch adds the `contabo_image_stats` data source. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 17 | The patch adds the `contabo_instances` data source. | `f7ccce5` | [PR #62](https://github.com/contabo/terraform-provider-contabo/pull/62), [Issue #54](https://github.com/contabo/terraform-provider-contabo/issues/54) | carried | v0.2.0 |
| 18 | The patch adds the `contabo_instance_available_products` data source. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 19 | The patch adds the `contabo_role` and `contabo_user` resources and the `contabo_role`, `contabo_api_permissions` and `contabo_users` data sources. | `6c294bc` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64) | carried | v0.2.0 |
| 20 | The patch lets you import a `contabo_object_storage` resource. The short region code and the region name now count as the same region. | `62d685a` | [PR #63](https://github.com/contabo/terraform-provider-contabo/pull/63), [Issue #34](https://github.com/contabo/terraform-provider-contabo/issues/34) | carried | v0.2.0 |
| 21 | The patch adds the `contabo_instance_vnc` data source, which reads the VNC settings of an instance, and the `contabo_instance_vnc` resource, which sets the VNC password. | `6c294bc`, `0b06d35` | [PR #64](https://github.com/contabo/terraform-provider-contabo/pull/64), [Issue #51](https://github.com/contabo/terraform-provider-contabo/issues/51) | carried. A test against the live API showed that the API ignores the `enabled` setting, so the API cannot turn VNC off. A strong VNC password is the best protection that the API offers. | v0.2.0 |
| 22 | The patch keeps one login token per account, renews the token during long runs, shows better API error messages, and reads the user id from the token. | `709a3f6` | [PR #60](https://github.com/contabo/terraform-provider-contabo/pull/60) | carried | v0.2.0 |
| 23 | The patch removes resources from the state when someone deleted them outside of Terraform, and waits for image downloads with a time limit. | `62d685a` | [PR #63](https://github.com/contabo/terraform-provider-contabo/pull/63) | carried | v0.2.0 |
| 24 | The patch updates terraform-plugin-sdk to version 2.40, sets up the release pipeline for `hayone1/contabo`, adds continuous integration, and adds unit tests against a fake API. | `a9cf95f`, `71fbcc9` | None. This patch only exists in the distribution. | carried | v0.2.0 |

## How to update the distribution from the official provider

Run these commands to bring the newest changes of the official provider into the distribution:

```sh
git fetch upstream
git merge upstream/main
```

Always use a merge and never a rebase. A rebase rewrites the history, and the published release tags must stay part of the history.

When the official provider has merged one of the patches, the merge often shows conflicts in the files of that patch. Contabo often squashes a pull request into one commit, which causes these conflicts. Resolve the conflicts in favor of the official provider. Then change the status of the patch in the table above to `merged upstream`, and delete the branch of the patch in the contrib fork.

## How to add a patch

The steps depend on where the patch should go:

- **A patch for the official provider:** In the clone of the contrib fork, create a branch with `git switch -c feature/<name> upstream/main`. Commit the change, push the branch, and open a pull request to the official repository. Then merge the branch into the distribution with `git fetch contrib && git merge --no-ff contrib/feature/<name>`.
- **A patch only for the distribution:** Commit the change on a `patch/<name>` branch in this repository, and merge the branch into `main`. Use this path also when Contabo declined the pull request.
- **A pull request that someone else opened in the official repository:** Fetch the pull request with `git fetch upstream pull/<n>/head:upstream-pr-<n>`, and merge the branch `upstream-pr-<n>` into `main`.
