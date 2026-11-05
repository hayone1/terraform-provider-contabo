# Publish an OpenID Connect discovery document in a public bucket.
resource "contabo_object_storage_bucket" "oidc" {
  name              = "my-cluster-oidc"
  object_storage_id = contabo_object_storage.this.id
  public_sharing    = true
}

resource "contabo_object_storage_object" "discovery" {
  object_storage_id = contabo_object_storage.this.id
  bucket            = contabo_object_storage_bucket.oidc.name
  key               = ".well-known/openid-configuration"
  content           = jsonencode({ issuer = contabo_object_storage_bucket.oidc.public_sharing_link })
  content_type      = "application/json"

  # Public buckets are served through a cache. Without this header, a changed
  # object can stay in the cache for a long time.
  cache_control = "public, max-age=60"
}
