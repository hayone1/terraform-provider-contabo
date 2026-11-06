# A script writes a new disk image over the running system of the instance.
# The old system cannot restart itself any more, so the provider restarts the
# instance through the Contabo API.
resource "terraform_data" "write_image" {
  # ... a provisioner that writes the disk image ...
}

resource "contabo_instance_restart" "boot_new_image" {
  instance_id = contabo_instance.node.id

  # Restart again whenever the image is written again.
  triggers = {
    image_written = terraform_data.write_image.id
  }
}
