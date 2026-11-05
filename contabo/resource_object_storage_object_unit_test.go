package contabo

import (
	"testing"

	"contabo.com/openapi"
)

func TestParseObjectStorageObjectId(t *testing.T) {
	cases := []struct {
		id                   string
		storage, bucket, key string
		wantErr              bool
	}{
		{id: "s/b/k", storage: "s", bucket: "b", key: "k"},
		{id: "s/b/.well-known/openid-configuration", storage: "s", bucket: "b", key: ".well-known/openid-configuration"},
		{id: "s/b/", wantErr: true},
		{id: "s/b", wantErr: true},
		{id: "/b/k", wantErr: true},
		{id: "s//k", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			storage, bucket, key, err := parseObjectStorageObjectId(c.id)
			if (err != nil) != c.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && (storage != c.storage || bucket != c.bucket || key != c.key) {
				t.Errorf("got %q %q %q", storage, bucket, key)
			}
		})
	}
}

func TestObjectPublicUrl(t *testing.T) {
	storage := openapi.ObjectStorageResponse{S3Url: "https://eu2.contabostorage.com", S3TenantId: "abc123"}
	got := objectPublicUrl(storage, "oidc", ".well-known/openid-configuration")
	want := "https://eu2.contabostorage.com/abc123:oidc/.well-known/openid-configuration"
	if got != want {
		t.Errorf("objectPublicUrl() = %q, want %q", got, want)
	}
	if got := objectPublicUrl(storage, "oidc", "a b/c"); got != "https://eu2.contabostorage.com/abc123:oidc/a%20b/c" {
		t.Errorf("objectPublicUrl() did not escape the key: %q", got)
	}
}

func TestValidateObjectKey(t *testing.T) {
	for key, wantErr := range map[string]bool{"": true, "/a": true, "a": false, ".well-known/x": false} {
		_, errs := validateObjectKey(key, "key")
		if (len(errs) > 0) != wantErr {
			t.Errorf("validateObjectKey(%q) errors = %v, wantErr %v", key, errs, wantErr)
		}
	}
}

func TestMd5Hex(t *testing.T) {
	if got := md5Hex([]byte(`{"test":1}` + "\n")); got != "4a45c31b88fef686732231fd53bd3546" {
		t.Errorf("md5Hex() = %q; this value is the ETag that Contabo returned for the same content", got)
	}
}
