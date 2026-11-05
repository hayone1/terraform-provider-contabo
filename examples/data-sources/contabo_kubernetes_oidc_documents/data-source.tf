# Compute the issuer documents of a Talos cluster from its service account key.
# The documents can be published before the cluster exists.
resource "talos_machine_secrets" "this" {}

data "contabo_kubernetes_oidc_documents" "this" {
  issuer          = "https://eu2.contabostorage.com/${data.contabo_object_storage.this.s3_tenant_id}:my-cluster-oidc"
  public_keys_pem = [base64decode(talos_machine_secrets.this.machine_secrets.certs.k8s_serviceaccount.key)]
}

resource "contabo_object_storage_object" "jwks" {
  object_storage_id = data.contabo_object_storage.this.id
  bucket            = "my-cluster-oidc"
  key               = "openid/v1/jwks"
  content           = data.contabo_kubernetes_oidc_documents.this.jwks
  content_type      = "application/json"
  cache_control     = "public, max-age=60"
}
