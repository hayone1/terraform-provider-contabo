# Changelog

This file lists every release of the `hayone1/contabo` distribution. The version numbers of the distribution do not follow the version numbers of the official provider. Each release names the version of the official provider that the release is based on.

## v0.2.0 (not released yet)

This release is based on version `v0.1.44` of the official provider. This release is the first release of the distribution.

### Behavior changes to check when you move from `contabo/contabo`

- `contabo_instance`: The provider now orders the add-ons in `add_ons` when the provider creates an instance. The official provider ignored the `add_ons` setting. When you add private networking (add-on `1477`) or automated backup (add-on `2214`) to an existing instance, the provider now orders the add-on through the upgrade API. When you add any other add-on to an existing instance, the plan fails, because the API cannot add that add-on. When you set `add_ons`, the state lists only the add-ons from your configuration.
- `contabo_instance`: When you change `product_id` or `region` of an existing instance, the plan now fails with an error. The official provider silently ignored such a change.
- `contabo_instance`: A reinstall now sends the values of `user_data`, `ssh_keys`, `root_password`, `default_user` and `application_id` from your configuration. The official provider sent only the values that had changed, so the reinstalled server lost the other values.
- `contabo_instance`: When you adopt an instance with `existing_instance_id`, the provider no longer reinstalls the instance.
- `contabo_instance`: The new setting `cancel_on_destroy` decides whether destroying the resource cancels the instance. The default value is `true`, which matches the old behavior.
- `contabo_firewall`: A rule without an `action` value is now created with the action `accept`, and a rule without a `status` value is now created with the status `active`. The official provider silently skipped such rules. A rule for every protocol shows the protocol `any`. The `rules` setting accepts one block. The `instances_status` attribute is now read-only. The provider no longer shows the "Block all traffic" rule that the API adds to the end of every firewall.
- `contabo_firewall` data source: The data source now finds a firewall by `id` or by `name`. All other attributes of the data source are read-only.
- `contabo_object_storage`: The `region` setting accepts the short region code (for example `EU`) and the region name (for example `European Union`). When you change the region of an existing object storage, the plan fails with an error.
- All resources: When someone deletes a resource outside of Terraform, the provider removes the resource from the state. The official provider failed the plan instead.
- The provider stores its login token in the folder `~/.cache/contabo/terraform/`. The provider now keeps one token file for each combination of token URL, client id and user. The environment variable `CNTB_TOKEN_CACHE_DIR` moves the token files to a different folder.

### Fixes

- The API client builds from the current API specification. This fix comes from pull request #57 of the official repository.
- The provider can read, import and create firewalls that protect instances. The API returns the list of instances as text that contains JSON (for example `"[203235786]"`), and the official provider could not read that text. This fix closes issues #52 and #56.
- The provider no longer crashes when a firewall rule has an empty `src_cidr` block. This fix closes issue #56.
- A reinstall keeps the `user_data` value (issue #49) and the `default_user` value (issues #45 and #47).
- The provider reports errors when a change of the display name or a reinstall of an instance fails.
- You can destroy an instance that you adopted with `existing_instance_id`. This fix closes issue #38.
- An imported object storage no longer shows changes in every plan. This fix closes issue #34.
- When you change the instances of a firewall, the provider now adds and removes only the changed instances. The official provider first removed every instance from the firewall, so the instances were unprotected for a moment.
- The provider waits for new instances and for image downloads with a time limit. The official provider waited without a limit.
- The provider no longer crashes when reading an image fails or when an object storage is missing.
- Error messages from the API now show the request and the real HTTP status code.
- During a long `terraform apply`, the provider logs in again when the login token expires.

### New resources

- `contabo_dns_zone`, `contabo_dns_record` and `contabo_ptr_record` manage DNS zones, DNS records and reverse DNS records. These resources come from pull request #59 of the official repository.
- `contabo_instance_reinstall` reinstalls an instance.
- `contabo_vip_assignment` assigns a virtual IP address to an instance.
- `contabo_role` manages the roles of your account.
- `contabo_user` manages the users of your account.
- `contabo_instance_vnc` sets the password of the VNC console of an instance. The API cannot turn the VNC console off, so a strong password is the best protection. This resource answers part of issue #51.

### New data sources

- `contabo_dns_zone` and `contabo_ptr_record` read DNS zones and reverse DNS records. These data sources come from pull request #59 of the official repository.
- `contabo_instances` lists instances. This data source closes issue #54.
- `contabo_vip` and `contabo_vips` read virtual IP addresses.
- `contabo_object_storage_credentials` reads the S3 keys of the API user.
- `contabo_image_stats` reads how much custom image storage is used and how much is free.
- `contabo_instance_available_products` lists the products that an instance can be upgraded to.
- `contabo_instance_vnc` reads the VNC settings of an instance. This data source answers part of issue #51.
- `contabo_role`, `contabo_api_permissions` and `contabo_users` read roles, API permissions and users.

### Other changes

- You can import a `contabo_object_storage_bucket` resource. This change comes from pull request #58 of the official repository.
- The `contabo_instance` resource has the new attributes `region_name`, `data_center`, `product_name` and `application_id`. The provider now reads the `default_user` value from the API.
- Instances, firewalls, reinstalls and images have time limits that you can change with a `timeouts` block.
- The provider uses version 2.40 of terraform-plugin-sdk.
