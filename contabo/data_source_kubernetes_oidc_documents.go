package contabo

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// This data source does not call the Contabo API. It computes the two
// documents that a Kubernetes API server serves for service account token
// discovery, so that the documents can be published in a public bucket before
// the cluster exists and without access to the cluster.
//
// The output matches kube-apiserver (pkg/serviceaccount/openidmetadata.go):
// the key id is the unpadded base64url SHA-256 hash of the PKIX DER encoding
// of the public key, RSA keys use RS256, and ECDSA keys use ES256/ES384/ES512.

func dataSourceKubernetesOidcDocuments() *schema.Resource {
	return &schema.Resource{
		Description: "This data source computes the OpenID Connect discovery document and the JSON Web Key Set (JWKS) that a Kubernetes API server serves for its service account tokens. " +
			"The data source does not call the Contabo API or the cluster. The documents are identical to the documents that the API server serves at `/.well-known/openid-configuration` and `/openid/v1/jwks`, " +
			"so you can publish them in a public bucket and use the bucket address as the service account issuer. Cloud providers such as Microsoft Entra ID, AWS IAM and Google Cloud then trust tokens that the cluster issues.",
		ReadContext: dataSourceKubernetesOidcDocumentsRead,
		Schema: map[string]*schema.Schema{
			"issuer": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The service account issuer of the cluster, which is the value of the API server flag `--service-account-issuer`. " +
					"The value must be an `https` address without a query, a fragment or a trailing slash.",
			},
			"jwks_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The address of the JWKS, which is the value of the API server flag `--service-account-jwks-uri`. The default value is `<issuer>/openid/v1/jwks`.",
			},
			"public_keys_pem": {
				Type:      schema.TypeList,
				Required:  true,
				MinItems:  1,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Description: "The keys that the API server uses to check service account tokens, in PEM format. A private key is accepted too; " +
					"the data source only uses its public part and never writes the private key to an output. Accepted formats are `RSA PRIVATE KEY`, `EC PRIVATE KEY`, `PRIVATE KEY`, `RSA PUBLIC KEY` and `PUBLIC KEY`. " +
					"List more than one key while you rotate the service account key.",
			},
			"openid_configuration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The discovery document as JSON. Publish it at `<issuer>/.well-known/openid-configuration`.",
			},
			"jwks": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The JSON Web Key Set as JSON. Publish it at the `jwks_uri` address.",
			},
			"key_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The key ids in the JWKS. Service account tokens carry one of these values in their `kid` header.",
			},
		},
	}
}

type jsonWebKey struct {
	Use string `json:"use"`
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Crv string `json:"crv,omitempty"`
	Alg string `json:"alg"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

type openIdMetadata struct {
	Issuer        string   `json:"issuer"`
	JwksUri       string   `json:"jwks_uri"`
	ResponseTypes []string `json:"response_types_supported"`
	SubjectTypes  []string `json:"subject_types_supported"`
	SigningAlgs   []string `json:"id_token_signing_alg_values_supported"`
}

func validateIssuerUrl(issuer string) error {
	u, err := url.Parse(issuer)
	if err != nil {
		return fmt.Errorf("the issuer %q is not a valid address: %w", issuer, err)
	}
	switch {
	case u.Scheme != "https":
		return fmt.Errorf("the issuer %q must use https", issuer)
	case u.Host == "":
		return fmt.Errorf("the issuer %q has no host", issuer)
	case u.RawQuery != "" || u.Fragment != "" || strings.HasSuffix(issuer, "?") || strings.HasSuffix(issuer, "#"):
		return fmt.Errorf("the issuer %q must not have a query or a fragment", issuer)
	case strings.HasSuffix(issuer, "/"):
		return fmt.Errorf("the issuer %q must not end with a slash", issuer)
	}
	return nil
}

// parsePublicKey returns the public key in a PEM block. Private keys are
// reduced to their public part.
func parsePublicKey(pemText string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(pemText)))
	if block == nil {
		return nil, fmt.Errorf("the value is not in PEM format")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &key.PublicKey, nil
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &key.PublicKey, nil
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("the private key type %T is not supported", key)
		}
		return signer.Public(), nil
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		return x509.ParsePKIXPublicKey(block.Bytes)
	}
	return nil, fmt.Errorf("the PEM type %q is not supported", block.Type)
}

func base64Url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// kubernetesKeyId computes the key id the way kube-apiserver does.
func kubernetesKeyId(publicKey crypto.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return base64Url(sum[:]), nil
}

func paddedCoordinate(value *big.Int, curve elliptic.Curve) []byte {
	size := (curve.Params().BitSize + 7) / 8
	out := make([]byte, size)
	value.FillBytes(out)
	return out
}

func jsonWebKeyFor(publicKey crypto.PublicKey) (jsonWebKey, error) {
	kid, err := kubernetesKeyId(publicKey)
	if err != nil {
		return jsonWebKey{}, err
	}
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		return jsonWebKey{
			Use: "sig", Kty: "RSA", Kid: kid, Alg: "RS256",
			N: base64Url(key.N.Bytes()),
			E: base64Url(big.NewInt(int64(key.E)).Bytes()),
		}, nil
	case *ecdsa.PublicKey:
		var alg, crv string
		switch key.Curve {
		case elliptic.P256():
			alg, crv = "ES256", "P-256"
		case elliptic.P384():
			alg, crv = "ES384", "P-384"
		case elliptic.P521():
			alg, crv = "ES512", "P-521"
		default:
			return jsonWebKey{}, fmt.Errorf("the elliptic curve %s is not supported", key.Curve.Params().Name)
		}
		return jsonWebKey{
			Use: "sig", Kty: "EC", Kid: kid, Crv: crv, Alg: alg,
			X: base64Url(paddedCoordinate(key.X, key.Curve)),
			Y: base64Url(paddedCoordinate(key.Y, key.Curve)),
		}, nil
	}
	return jsonWebKey{}, fmt.Errorf("the key type %T is not supported; use an RSA or an ECDSA key", publicKey)
}

// kubernetesOidcDocuments builds the discovery document and the JWKS.
func kubernetesOidcDocuments(issuer string, jwksUri string, pems []string) (string, string, []string, error) {
	if err := validateIssuerUrl(issuer); err != nil {
		return "", "", nil, err
	}
	if jwksUri == "" {
		jwksUri = issuer + "/openid/v1/jwks"
	}
	if u, err := url.Parse(jwksUri); err != nil || u.Scheme != "https" || u.Host == "" {
		return "", "", nil, fmt.Errorf("the jwks_uri %q must be an https address", jwksUri)
	}

	keys := []jsonWebKey{}
	seen := map[string]bool{}
	algs := map[string]bool{}
	for i, pemText := range pems {
		publicKey, err := parsePublicKey(pemText)
		if err != nil {
			return "", "", nil, fmt.Errorf("public_keys_pem[%d]: %w", i, err)
		}
		jwk, err := jsonWebKeyFor(publicKey)
		if err != nil {
			return "", "", nil, fmt.Errorf("public_keys_pem[%d]: %w", i, err)
		}
		if seen[jwk.Kid] {
			continue
		}
		seen[jwk.Kid] = true
		algs[jwk.Alg] = true
		keys = append(keys, jwk)
	}

	signingAlgs := make([]string, 0, len(algs))
	for alg := range algs {
		signingAlgs = append(signingAlgs, alg)
	}
	sort.Strings(signingAlgs)

	metadata, err := json.Marshal(openIdMetadata{
		Issuer:        issuer,
		JwksUri:       jwksUri,
		ResponseTypes: []string{"id_token"},
		SubjectTypes:  []string{"public"},
		SigningAlgs:   signingAlgs,
	})
	if err != nil {
		return "", "", nil, err
	}
	jwks, err := json.Marshal(map[string][]jsonWebKey{"keys": keys})
	if err != nil {
		return "", "", nil, err
	}
	keyIds := make([]string, len(keys))
	for i, key := range keys {
		keyIds[i] = key.Kid
	}
	return string(metadata), string(jwks), keyIds, nil
}

func dataSourceKubernetesOidcDocumentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	issuer := d.Get("issuer").(string)
	jwksUri := d.Get("jwks_uri").(string)
	var pems []string
	for _, v := range d.Get("public_keys_pem").([]interface{}) {
		s, _ := v.(string)
		pems = append(pems, s)
	}
	metadata, jwks, keyIds, err := kubernetesOidcDocuments(issuer, jwksUri, pems)
	if err != nil {
		return diag.FromErr(err)
	}
	if jwksUri == "" {
		jwksUri = issuer + "/openid/v1/jwks"
	}
	sum := sha256.Sum256([]byte(metadata + jwks))
	d.SetId(base64Url(sum[:]))
	for name, value := range map[string]interface{}{
		"jwks_uri":             jwksUri,
		"openid_configuration": metadata,
		"jwks":                 jwks,
		"key_ids":              keyIds,
	} {
		if err := d.Set(name, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
