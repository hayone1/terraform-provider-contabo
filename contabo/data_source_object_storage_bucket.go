package contabo

import (
	"context"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceObjectStorageBucket() *schema.Resource {
	return &schema.Resource{
		Description: "Manage buckets on your contabo Object Storage. With this resource you are able to manage your buckets the same way your are able to manage them in your contabo customer panel.",
		ReadContext: dataSourceObjectStorageBucketRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of your bucket, consider the naming restriction https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-s3-bucket-naming-requirements.html.",
				Optional:    true,
			},
			"object_storage_id": {
				Type:        schema.TypeString,
				Description: "The contabo objectStorageId on which the bucket should be created.",
				Required:    true,
			},
			"public_sharing": {
				Type:        schema.TypeBool,
				Description: "Whether the bucket is shared publicly. The data source reads the value from the bucket policy. A value in the configuration is ignored; the setting only stays optional for compatibility.",
				Optional:    true,
				Computed:    true,
			},
			"public_sharing_link": {
				Type:        schema.TypeString,
				Description: "The public address of the bucket, `<s3 url>/<s3 tenant id>:<bucket>`. The value is empty when the bucket is not shared publicly.",
				Computed:    true,
			},
			"creation_date": {
				Type:        schema.TypeString,
				Description: "The creation date of the bucket.",
				Computed:    true,
			},
		},
	}
}

func dataSourceObjectStorageBucketRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*openapi.APIClient)
	bucketName := d.Get("name").(string)
	objectStorageIdOfBucket := d.Get("object_storage_id").(string)

	diags, objectStorage, s3Credentials := getObjectStorageAndCredentials(diags, client, objectStorageIdOfBucket)
	if diags.HasError() {
		return diags
	}

	diags, bucketInfo, err := getBucket(diags, objectStorage, s3Credentials, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}
	if bucketInfo.Name == "" {
		return diag.Errorf("the bucket %q does not exist in the object storage %s", bucketName, objectStorageIdOfBucket)
	}

	// Read the real sharing state. Earlier versions returned the configured
	// value and an empty link instead.
	diags, isPublicSharing, err := isBucketPublicSharing(diags, objectStorage, s3Credentials, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}
	publicSharingLink := ""
	if isPublicSharing {
		publicSharingLink = generatePublicSharingUrl(objectStorage, bucketName)
	}

	// The shared AddObjectStorageBucketToData helper sets an "id" attribute,
	// which only the resource schema has, so the data source sets its values
	// itself.
	d.SetId(generateBucketId(objectStorageIdOfBucket, bucketName))
	for name, value := range map[string]interface{}{
		"name":                bucketName,
		"object_storage_id":   objectStorageIdOfBucket,
		"creation_date":       bucketInfo.CreationDate.String(),
		"public_sharing":      isPublicSharing,
		"public_sharing_link": publicSharingLink,
	} {
		if err := d.Set(name, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}
