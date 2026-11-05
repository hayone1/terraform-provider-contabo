package contabo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
)

func pemOf(t *testing.T, kind string, der []byte) string {
	t.Helper()
	return string(pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der}))
}

type testJwks struct {
	Keys []map[string]string `json:"keys"`
}

func TestKubernetesOidcDocumentsRsa(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privatePem := pemOf(t, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
	issuer := "https://eu2.contabostorage.com/abc:my-cluster-oidc"

	metadata, jwksText, keyIds, err := kubernetesOidcDocuments(issuer, "", []string{privatePem})
	if err != nil {
		t.Fatal(err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &meta); err != nil {
		t.Fatal(err)
	}
	if meta["issuer"] != issuer || meta["jwks_uri"] != issuer+"/openid/v1/jwks" {
		t.Errorf("unexpected discovery document: %s", metadata)
	}
	if algs := meta["id_token_signing_alg_values_supported"].([]interface{}); len(algs) != 1 || algs[0] != "RS256" {
		t.Errorf("unexpected algorithms: %v", algs)
	}

	// The key id must match kube-apiserver: base64url(sha256(PKIX DER)).
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	sum := sha256.Sum256(der)
	wantKid := base64.RawURLEncoding.EncodeToString(sum[:])
	if len(keyIds) != 1 || keyIds[0] != wantKid {
		t.Errorf("key ids = %v, want [%s]", keyIds, wantKid)
	}

	var jwks testJwks
	if err := json.Unmarshal([]byte(jwksText), &jwks); err != nil {
		t.Fatal(err)
	}
	jwk := jwks.Keys[0]
	n, _ := base64.RawURLEncoding.DecodeString(jwk["n"])
	if new(big.Int).SetBytes(n).Cmp(key.N) != 0 || jwk["e"] != "AQAB" || jwk["kid"] != wantKid || jwk["alg"] != "RS256" || jwk["use"] != "sig" {
		t.Errorf("unexpected JWK: %v", jwk)
	}
	if strings.Contains(jwksText, "PRIVATE") || strings.Contains(jwksText, `"d"`) {
		t.Error("the JWKS contains private key material")
	}

	// The public key in every accepted format gives the same document.
	for _, alt := range []string{
		pemOf(t, "RSA PUBLIC KEY", x509.MarshalPKCS1PublicKey(&key.PublicKey)),
		pemOf(t, "PUBLIC KEY", der),
	} {
		_, other, _, err := kubernetesOidcDocuments(issuer, "", []string{alt})
		if err != nil || other != jwksText {
			t.Errorf("different JWKS for %q: %v", strings.SplitN(alt, "\n", 2)[0], err)
		}
	}

	// A repeated key is listed once.
	_, _, keyIds, _ = kubernetesOidcDocuments(issuer, "", []string{privatePem, pemOf(t, "PUBLIC KEY", der)})
	if len(keyIds) != 1 {
		t.Errorf("a repeated key was listed %d times", len(keyIds))
	}
}

func TestKubernetesOidcDocumentsEcdsaAndRotation(t *testing.T) {
	ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecDer, _ := x509.MarshalECPrivateKey(ec)
	rs, _ := rsa.GenerateKey(rand.Reader, 2048)
	pkcs8, _ := x509.MarshalPKCS8PrivateKey(rs)

	metadata, jwksText, keyIds, err := kubernetesOidcDocuments("https://issuer.example.com", "https://keys.example.com/jwks",
		[]string{pemOf(t, "EC PRIVATE KEY", ecDer), pemOf(t, "PRIVATE KEY", pkcs8)})
	if err != nil {
		t.Fatal(err)
	}
	if len(keyIds) != 2 {
		t.Fatalf("key ids = %v", keyIds)
	}
	if !strings.Contains(metadata, `"jwks_uri":"https://keys.example.com/jwks"`) || !strings.Contains(metadata, `["ES256","RS256"]`) {
		t.Errorf("unexpected discovery document: %s", metadata)
	}
	var jwks testJwks
	_ = json.Unmarshal([]byte(jwksText), &jwks)
	x, _ := base64.RawURLEncoding.DecodeString(jwks.Keys[0]["x"])
	if jwks.Keys[0]["crv"] != "P-256" || len(x) != 32 {
		t.Errorf("unexpected EC key: %v", jwks.Keys[0])
	}
}

func TestValidateIssuerUrl(t *testing.T) {
	for issuer, wantErr := range map[string]bool{
		"https://eu2.contabostorage.com/abc:bucket": false,
		"https://issuer.example.com":                false,
		"http://issuer.example.com":                 true,
		"https://issuer.example.com/":               true,
		"https://issuer.example.com?x=1":            true,
		"https://issuer.example.com#a":              true,
		"https://":                                  true,
		"not a url":                                 true,
	} {
		if err := validateIssuerUrl(issuer); (err != nil) != wantErr {
			t.Errorf("validateIssuerUrl(%q) = %v, wantErr %v", issuer, err, wantErr)
		}
	}
}

func TestKubernetesOidcDocumentsBadInput(t *testing.T) {
	if _, _, _, err := kubernetesOidcDocuments("https://i.example.com", "", []string{"not pem"}); err == nil {
		t.Error("expected an error for a value that is not PEM")
	}
	if _, _, _, err := kubernetesOidcDocuments("https://i.example.com", "http://plain", []string{}); err == nil {
		t.Error("expected an error for an http jwks_uri")
	}
}
