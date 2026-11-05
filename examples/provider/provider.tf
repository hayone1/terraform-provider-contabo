terraform {
  required_providers {
    contabo = {
      source  = "hayone1/contabo"
      version = ">= 0.2.0"
    }
  }
}

# Configure your Contabo API credentials in the provider block or with the
# CNTB_OAUTH2_CLIENT_ID, CNTB_OAUTH2_CLIENT_SECRET, CNTB_OAUTH2_USER and
# CNTB_OAUTH2_PASS environment variables.
provider "contabo" {
  oauth2_client_id     = "[your client id]"
  oauth2_client_secret = "[your client secret]"
  oauth2_user          = "[your username]"
  oauth2_pass          = "[your password]"
}

# Create a default contabo VPS instance
resource "contabo_instance" "default_instance" {}

# Output our newly created instances
output "default_instance_output" {
  description = "Our first default instance"
  value       = contabo_instance.default_instance
}
