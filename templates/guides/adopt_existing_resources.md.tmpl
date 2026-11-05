---
subcategory: ""
page_title: "Manage existing resources"
description: |-
    Bring instances and other resources that were ordered in the customer control panel under Terraform without reinstalling or cancelling them.
---

# Manage existing resources

Resources ordered in the customer control panel can be managed by Terraform. Instances need care, because changing an instance can reinstall it and destroying it cancels it.

## Find the resources

The `contabo_instances` data source lists instances by display name, product, region, status or tag:

```terraform
data "contabo_instances" "workers" {
  display_name_prefix = "worker-"
}

output "worker_ids" {
  value = data.contabo_instances.workers.ids
}
```

## Import

Import blocks (Terraform 1.5 and later) or `terraform import` add a resource to the state without changing it:

```terraform
import {
  to = contabo_instance.worker
  id = "203235629"
}

resource "contabo_instance" "worker" {
  cancel_on_destroy = false
}
```

Run `terraform plan -generate-config-out=generated.tf` to let Terraform write the configuration, then review the plan before you apply it. A plan for an imported instance should show no changes.

These attributes reinstall an instance when they change: `image_id`, `ssh_keys`, `root_password`, `user_data`, `default_user` and `application_id`. The API does not return `user_data` and `root_password`, so set them only when you intend to reinstall, or add them to `lifecycle.ignore_changes`.

## Keep instances when the resource is destroyed

`cancel_on_destroy = false` makes `terraform destroy` remove the instance from the state only. The instance keeps running and billing. Use it for instances that Terraform did not order, or that must survive a mistaken destroy.

## `existing_instance_id`

`existing_instance_id` adopts an instance during `terraform apply` instead of an import. The provider applies `display_name` and add-ons but does not reinstall the instance during adoption. If the configuration differs from the instance in a reinstall attribute, the next plan reinstalls it, so match those attributes to the instance or ignore them.

## Moving from the official provider

See the migration steps on the [provider page](../index.md#migrating-from-contabocontabo).
