package contabo

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInstanceRestart() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource restarts an existing instance through the Contabo API. The provider restarts the instance when Terraform creates this resource, and again whenever a value in `triggers` changes. When Terraform destroys this resource, nothing happens to the instance.\n\nThe restart also works when the operating system of the instance no longer responds, for example after a program wrote a new disk image over the running system. The provider waits until Contabo reports the instance as running again.",
		CreateContext: resourceInstanceRestartCreate,
		ReadContext:   resourceInstanceRestartRead,
		DeleteContext: schema.NoopContext,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the instance that the provider restarts.",
			},
			"triggers": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A map of any values that you choose. When a value in this map changes, the provider restarts the instance again.",
			},
			"restarted_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time of the restart in RFC 3339 format.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the instance after the restart.",
			},
		},
	}
}

func resourceInstanceRestartCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(d.Get("instance_id").(string), 10, 64)
	if err != nil {
		return diag.Errorf("instance_id must be a numeric instance id: %s", err)
	}

	httpResp, err := apiRequest(ctx, client, http.MethodPost, instancePath(instanceId)+"/actions/restart", nil, map[string]interface{}{}, nil)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	if _, diags := waitForInstanceReady(ctx, client, instanceId, d.Timeout(schema.TimeoutCreate)); diags.HasError() {
		return diags
	}

	restartedAt := time.Now().UTC().Format(time.RFC3339)
	d.SetId(fmt.Sprintf("%d/%s", instanceId, restartedAt))
	_ = d.Set("restarted_at", restartedAt)
	return resourceInstanceRestartRead(ctx, d, m)
}

func resourceInstanceRestartRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
