package contabo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceInstanceReinstall() *schema.Resource {
	return &schema.Resource{
		Description:   "Reinstalls an existing instance. The reinstall runs when the resource is created and again whenever an argument or a value in `triggers` changes, for example after a private network was attached and the instance needs a fresh installation to configure it. Destroying the resource does nothing to the instance.\n\nCAUTION: a reinstall erases the disk of the instance.\n\nWhen this resource manages the installation, add `image_id`, `ssh_keys`, `root_password`, `user_data`, `default_user` and `application_id` of the `contabo_instance` to its `lifecycle.ignore_changes`, so that the two resources do not reinstall in turns.",
		CreateContext: resourceInstanceReinstallCreate,
		ReadContext:   resourceInstanceReinstallRead,
		DeleteContext: schema.NoopContext,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the instance to reinstall.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the image to install.",
			},
			"ssh_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Secret ids of the public SSH keys for the default user.",
			},
			"root_password": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Secret id of the root password.",
			},
			"user_data": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Cloud-init configuration.",
			},
			"default_user": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"root", "admin", "administrator"}, false),
				Description:  "Default user: `root`, `admin` or, on Windows, `administrator`. The API defaults to `admin`.",
			},
			"application_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Id of an application to install with the image.",
			},
			"triggers": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Arbitrary values; any change reinstalls the instance.",
			},
			"reinstalled_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time of the reinstall, RFC 3339.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the instance after the reinstall.",
			},
		},
	}
}

func resourceInstanceReinstallCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(d.Get("instance_id").(string), 10, 64)
	if err != nil {
		return diag.Errorf("instance_id must be a numeric instance id: %s", err)
	}

	request := map[string]interface{}{"imageId": d.Get("image_id").(string)}
	setIfNotEmpty(request, "userData", d.Get("user_data").(string))
	setIfNotEmpty(request, "defaultUser", d.Get("default_user").(string))
	setIfNotEmpty(request, "applicationId", d.Get("application_id").(string))
	if sshKeys := interfacesToInt64s(d.Get("ssh_keys").([]interface{})); len(sshKeys) > 0 {
		request["sshKeys"] = sshKeys
	}
	if rootPassword := d.Get("root_password").(int); rootPassword > 0 {
		request["rootPassword"] = rootPassword
	}

	if diags := reinstallInstance(ctx, client, instanceId, request, d.Timeout(schema.TimeoutCreate)); diags.HasError() {
		return diags
	}
	reinstalledAt := time.Now().UTC().Format(time.RFC3339)
	d.SetId(fmt.Sprintf("%d/%s", instanceId, reinstalledAt))
	_ = d.Set("reinstalled_at", reinstalledAt)
	return resourceInstanceReinstallRead(ctx, d, m)
}

func resourceInstanceReinstallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(d.Get("instance_id").(string), 10, 64)
	if err != nil {
		return diag.Errorf("instance_id must be a numeric instance id: %s", err)
	}
	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		if removeFromStateIfNotFound(d, httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}
	if err := d.Set("status", instance.Status); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
