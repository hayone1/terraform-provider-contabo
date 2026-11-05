---
subcategory: ""
page_title: "Manage existing resources"
description: |-
    This guide shows how Terraform can manage instances and other resources that you ordered in the Customer Control Panel, without reinstalling or cancelling the instances.
---

# Manage existing resources

Terraform can manage resources that you ordered in the Customer Control Panel. Instances need extra care for two reasons. First, some changes to an instance reinstall the instance, and a reinstall erases the disk of the instance. Second, destroying an instance in Terraform cancels the instance at Contabo.

## Find the resources

The `contabo_instances` data source lists your instances. The data source can filter the instances by display name, product, region, status or tag. This example finds every instance whose display name starts with `worker-`:

```terraform
data "contabo_instances" "workers" {
  display_name_prefix = "worker-"
}

output "worker_ids" {
  value = data.contabo_instances.workers.ids
}
```

## Import the resources

An import adds an existing resource to the Terraform state. An import does not change the resource itself. You can use an `import` block, which needs Terraform 1.5 or newer, or the `terraform import` command. This example imports one instance:

```terraform
import {
  to = contabo_instance.worker
  id = "203235629"
}

resource "contabo_instance" "worker" {
  cancel_on_destroy = false
}
```

Terraform can write the configuration of the imported resource for you. Run `terraform plan -generate-config-out=generated.tf`, and Terraform writes the configuration into the file `generated.tf`. Read the plan before you apply the plan. The plan for an imported instance should show no changes.

A change to one of these instance settings reinstalls the instance: `image_id`, `ssh_keys`, `root_password`, `user_data`, `default_user` and `application_id`. The API does not return the `user_data` value or the `root_password` value of an instance. For that reason, set `user_data` and `root_password` only when you want to reinstall the instance. You can also add these settings to `lifecycle.ignore_changes`, so that Terraform ignores changes to them.

## Keep an instance when Terraform destroys the resource

When you set `cancel_on_destroy = false` on an instance, `terraform destroy` only removes the instance from the Terraform state. The instance keeps running, and Contabo keeps billing you for the instance. Use this setting for instances that Terraform did not order. Use this setting also for instances that must survive a destroy by mistake.

## The `existing_instance_id` setting

The `existing_instance_id` setting is a second way to adopt an instance. With this setting, the provider adopts the instance during `terraform apply`, instead of during an import. During the adoption, the provider sets the display name and orders the add-ons from your configuration, but the provider does not reinstall the instance.

After the adoption, the next plan compares your configuration with the instance. If a reinstall setting in your configuration is different from the value on the instance, the next plan reinstalls the instance. To avoid a reinstall, copy the values from the instance into your configuration, or add the settings to `lifecycle.ignore_changes`.

## Moving from the official provider

The [provider page](../index.md#moving-from-contabocontabo) explains how to move your Terraform state from the official provider to this distribution.
