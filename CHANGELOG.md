# Changelog

All releases of the `hayone1/contabo` distribution. Version numbers are independent of upstream; each release names its upstream base.

## v0.2.0 (unreleased)

Based on upstream `v0.1.44`. First release of the distribution.

### Behavior changes to check when migrating from `contabo/contabo`

- `contabo_instance`: `add_ons` are now ordered when the instance is created. They were ignored before. Adding private networking (`1477`) or backup (`2214`) to an existing instance now orders the add-on through the upgrade API; other added add-ons fail at plan time. When `add_ons` is set, the state lists only the configured add-ons.
- `contabo_instance`: changing `product_id` or `region` of an existing instance now fails at plan time. It was silently ignored.
- `contabo_instance`: a reinstall now sends `user_data`, `ssh_keys`, `root_password`, `default_user` and `application_id` from the configuration, not only the changed ones.
- `contabo_instance`: adopting an instance with `existing_instance_id` no longer reinstalls it.
- `contabo_instance`: new `cancel_on_destroy` (default `true`, as before).
- `contabo_firewall`: rules without `action` or `status` are now created as `accept` and `active`; before, they were silently dropped. `protocol = "any"` is reported as `any`. `rules` accepts one block. `instances_status` is read only. The drop-all rule that the API appends is no longer shown.
- `contabo_firewall` data source: looks the firewall up by `id` or `name`; all other attributes are read only.
- `contabo_object_storage`: `region` accepts the slug (`EU`) and the name (`European Union`); changing it fails at plan time.
- Resources deleted outside Terraform are removed from the state instead of failing the plan.
- The OAuth2 token cache in `~/.cache/contabo/terraform/` is now one file per token URL, client id and user. `CNTB_TOKEN_CACHE_DIR` moves it.

### Fixes

- The API client builds from the current API spec (upstream PR #57).
- Firewalls with instances can be read, imported and created: the API returns `instances` as a JSON-encoded string (issues #52, #56).
- Fixed a panic for firewall rules with an empty `src_cidr` block (issue #56).
- Reinstalls keep `user_data` (issue #49) and `default_user` (issues #45, #47).
- Errors from instance display name updates and reinstalls are reported.
- Instances adopted with `existing_instance_id` can be destroyed (issue #38).
- Object storage imports converge (issue #34).
- Firewall instance assignments change only the difference instead of unassigning every instance first.
- Waiting for instances and image downloads has a timeout instead of an unbounded recursive poll.
- Panics on failed image reads and on missing object storages are fixed.
- API errors show the request and the real status code.
- Long applies log in again when the refresh token has expired.

### New resources

- `contabo_dns_zone`, `contabo_dns_record`, `contabo_ptr_record` (upstream PR #59)
- `contabo_instance_reinstall`
- `contabo_vip_assignment`
- `contabo_role`
- `contabo_user`

### New data sources

- `contabo_dns_zone`, `contabo_ptr_record` (upstream PR #59)
- `contabo_instances` (issue #54)
- `contabo_vip`, `contabo_vips`
- `contabo_object_storage_credentials`
- `contabo_image_stats`
- `contabo_instance_available_products`
- `contabo_instance_vnc` (issue #51, read only)
- `contabo_role`, `contabo_api_permissions`, `contabo_users`

### Other

- Import for `contabo_object_storage_bucket` (upstream PR #58).
- New instance attributes: `region_name`, `data_center`, `product_name`, `application_id`; `default_user` is read from the API.
- Timeouts for instances, firewalls, reinstalls and images.
- terraform-plugin-sdk v2.40.
