package contabo

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// importContentLimit is the largest object that an import reads into the
// state. Larger objects cannot be imported, because the state would grow too
// much.
const importContentLimit = 1 << 20

func resourceObjectStorageObject() *schema.Resource {
	return &schema.Resource{
		Description: "This resource uploads a small object, such as a JSON document, into a bucket of your object storage. " +
			"The provider reads the S3 keys of its API user from the Contabo API, so you do not need an S3 provider or S3 keys in your configuration.\n\n" +
			"Contabo serves public buckets through a cache. An object without a `cache_control` value can stay in that cache for a long time after you change the object, and you cannot empty the cache. " +
			"Set `cache_control` on every object that a public bucket serves and that you plan to change, for example `public, max-age=60`.",
		CreateContext: resourceObjectStorageObjectPut,
		ReadContext:   resourceObjectStorageObjectRead,
		UpdateContext: resourceObjectStorageObjectPut,
		DeleteContext: resourceObjectStorageObjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceObjectStorageObjectImport,
		},
		Schema: map[string]*schema.Schema{
			"object_storage_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the object storage that holds the bucket.",
			},
			"bucket": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the bucket.",
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateObjectKey,
				Description:  "The key (the path) of the object inside the bucket, for example `.well-known/openid-configuration`. The key must not start with a slash.",
			},
			"content": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"content", "content_base64"},
				Description:  "The content of the object as text. Set either `content` or `content_base64`.",
			},
			"content_base64": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"content", "content_base64"},
				Description:  "The content of the object, encoded with base64. Use this setting for binary content. Set either `content` or `content_base64`.",
			},
			"content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "application/octet-stream",
				Description: "The media type of the object, for example `application/json`. The default value is `application/octet-stream`.",
			},
			"cache_control": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The `Cache-Control` header of the object, for example `public, max-age=60`. The cache in front of public buckets respects this value.",
			},
			"etag": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The entity tag of the object. For objects that this resource uploads, the value is the MD5 hash of the content.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public address of the object. The address only works when the bucket has `public_sharing` turned on.",
			},
		},
	}
}

func validateObjectKey(v interface{}, k string) ([]string, []error) {
	key := v.(string)
	if key == "" {
		return nil, []error{fmt.Errorf("%s must not be empty", k)}
	}
	if strings.HasPrefix(key, "/") {
		return nil, []error{fmt.Errorf("%s must not start with a slash: %q", k, key)}
	}
	return nil, nil
}

// objectContent returns the bytes that the configuration wants in the object.
func objectContent(d *schema.ResourceData) ([]byte, error) {
	if v, ok := d.GetOk("content_base64"); ok {
		data, err := base64.StdEncoding.DecodeString(v.(string))
		if err != nil {
			return nil, fmt.Errorf("content_base64 is not valid base64: %w", err)
		}
		return data, nil
	}
	return []byte(d.Get("content").(string)), nil
}

func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// objectPublicUrl builds the public address of an object in a publicly
// shared bucket: <s3 url>/<s3 tenant id>:<bucket>/<key>.
func objectPublicUrl(objectStorage openapi.ObjectStorageResponse, bucket string, key string) string {
	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return generatePublicSharingUrl(objectStorage, bucket) + "/" + strings.Join(segments, "/")
}

func newObjectStorageS3Client(objectStorage openapi.ObjectStorageResponse, s3Credentials S3Credentials) (*minio.Client, error) {
	s3Url, err := url.Parse(objectStorage.S3Url)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 URL %q: %w", objectStorage.S3Url, err)
	}
	return minio.New(s3Url.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Credentials.AccessKey, s3Credentials.SecretKey, ""),
		Secure: s3Url.Scheme != "http",
	})
}

// objectStorageS3Client looks up the object storage and the S3 keys of the
// provider's API user, and returns an S3 client for the object storage.
func objectStorageS3Client(client *openapi.APIClient, objectStorageId string) (*minio.Client, openapi.ObjectStorageResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	diags, objectStorage, s3Credentials := getObjectStorageAndCredentials(diags, client, objectStorageId)
	if diags.HasError() {
		return nil, objectStorage, diags
	}
	s3Client, err := newObjectStorageS3Client(objectStorage, s3Credentials)
	if err != nil {
		return nil, objectStorage, diag.FromErr(err)
	}
	return s3Client, objectStorage, nil
}

func objectStorageObjectId(objectStorageId string, bucket string, key string) string {
	return objectStorageId + "/" + bucket + "/" + key
}

// parseObjectStorageObjectId splits "<object storage id>/<bucket>/<key>". The
// key itself may contain slashes.
func parseObjectStorageObjectId(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("the id %q must have the form <object_storage_id>/<bucket>/<key>", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func resourceObjectStorageObjectPut(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	objectStorageId := d.Get("object_storage_id").(string)
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	data, err := objectContent(d)
	if err != nil {
		return diag.FromErr(err)
	}
	s3Client, _, diags := objectStorageS3Client(client, objectStorageId)
	if diags.HasError() {
		return diags
	}
	_, err = s3Client.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType:    d.Get("content_type").(string),
		CacheControl:   d.Get("cache_control").(string),
		SendContentMd5: true,
	})
	if err != nil {
		return diag.Errorf("uploading the object %q to the bucket %q failed: %s", key, bucket, err)
	}
	d.SetId(objectStorageObjectId(objectStorageId, bucket, key))
	return resourceObjectStorageObjectRead(ctx, d, m)
}

func isNoSuchKey(err error) bool {
	code := minio.ToErrorResponse(err).Code
	return code == "NoSuchKey" || code == "NoSuchBucket" || code == "NotFound"
}

func resourceObjectStorageObjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	objectStorageId, bucket, key, err := parseObjectStorageObjectId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	s3Client, objectStorage, diags := objectStorageS3Client(client, objectStorageId)
	if diags.HasError() {
		return diags
	}
	info, err := s3Client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if isNoSuchKey(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("reading the object %q in the bucket %q failed: %s", key, bucket, err)
	}

	etag := strings.Trim(info.ETag, `"`)
	// The ETag of an object that was uploaded in one request is the MD5 hash
	// of the content. A different ETag means that someone changed the object
	// outside of Terraform. Clearing the content makes the next plan upload
	// the configured content again.
	if want, err := objectContent(d); err == nil && d.Get("etag").(string) != "" && etag != md5Hex(want) {
		if _, ok := d.GetOk("content_base64"); ok {
			_ = d.Set("content_base64", "")
		} else {
			_ = d.Set("content", "")
		}
	}

	values := map[string]interface{}{
		"object_storage_id": objectStorageId,
		"bucket":            bucket,
		"key":               key,
		"etag":              etag,
		"url":               objectPublicUrl(objectStorage, bucket, key),
		"content_type":      info.ContentType,
		"cache_control":     info.Metadata.Get("Cache-Control"),
	}
	for name, value := range values {
		if err := d.Set(name, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceObjectStorageObjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	objectStorageId, bucket, key, err := parseObjectStorageObjectId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	s3Client, _, diags := objectStorageS3Client(client, objectStorageId)
	if diags.HasError() {
		return diags
	}
	if err := s3Client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{}); err != nil && !isNoSuchKey(err) {
		return diag.Errorf("deleting the object %q in the bucket %q failed: %s", key, bucket, err)
	}
	return nil
}

func resourceObjectStorageObjectImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	objectStorageId, bucket, key, err := parseObjectStorageObjectId(d.Id())
	if err != nil {
		return nil, err
	}
	client := m.(*openapi.APIClient)
	s3Client, _, diags := objectStorageS3Client(client, objectStorageId)
	if diags.HasError() {
		return nil, fmt.Errorf("%s", diags[0].Summary)
	}
	object, err := s3Client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = object.Close() }()
	data, err := io.ReadAll(io.LimitReader(object, importContentLimit+1))
	if err != nil {
		return nil, fmt.Errorf("reading the object %q in the bucket %q failed: %w", key, bucket, err)
	}
	if len(data) > importContentLimit {
		return nil, fmt.Errorf("the object %q is larger than %d bytes, and this resource only imports smaller objects", key, importContentLimit)
	}
	if utf8.Valid(data) {
		_ = d.Set("content", string(data))
	} else {
		_ = d.Set("content_base64", base64.StdEncoding.EncodeToString(data))
	}
	_ = d.Set("object_storage_id", objectStorageId)
	_ = d.Set("bucket", bucket)
	_ = d.Set("key", key)
	return []*schema.ResourceData{d}, nil
}
