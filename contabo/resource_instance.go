package contabo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// instanceUpgradeAddOns maps the add-on ids that can be added to an existing
// instance to the attribute of the upgrade request that orders them. The ids
// are the catalog group ids that the API reports in addOns[].id.
var instanceUpgradeAddOns = map[int64]string{
	1477: "privateNetworking",
	2214: "backup",
}

// Attributes whose change reinstalls the instance.
var instanceReinstallAttributes = []string{
	"image_id", "ssh_keys", "root_password", "user_data", "default_user", "application_id",
}

func instanceIpSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Address",
			},
			"netmask_cidr": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Netmask CIDR",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway",
			},
		},
	}
}

func instanceIpConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Primary IPv4 and IPv6 configuration of the instance.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"v4": {Type: schema.TypeList, Computed: true, Elem: instanceIpSchema()},
				"v6": {Type: schema.TypeList, Computed: true, Elem: instanceIpSchema()},
			},
		},
	}
}

func instanceAdditionalIpsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "All other additional IP addresses of the instance.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"v4": {Type: schema.TypeList, Computed: true, Elem: instanceIpSchema()},
			},
		},
	}
}

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a compute instance (VPS or VDS). Creating an instance orders it and starts billing; destroying it cancels it at the end of the contract period, unless `cancel_on_destroy` is false. It also supports [cloud-init](https://cloud-init.io/) at least on the default images (custom images need cloud-init support packages) through the `user_data` field.",
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		CustomizeDiff: resourceInstanceCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the compute instance. Use it to manage it!",
			},
			"existing_instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Id of an existing instance to manage instead of ordering a new one. The provider adopts the instance as it is: it applies `display_name` and add-ons but does not reinstall it on create. Changes to reinstall attributes on later applies reinstall it as usual. Consider `terraform import` instead, and set `cancel_on_destroy = false` if destroying the resource should not cancel the instance.",
			},
			"cancel_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether destroying the resource cancels the instance. When false, destroy only removes the instance from the Terraform state and the instance keeps running and billing. Defaults to `true`.",
			},
			"last_updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time of the last update of the compute instance.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the compute instance.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The instance name chosen by the customer that will be shown in the customer panel.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "CAUTION: On updating this value your server will be reinstalled! Image Id is used to set up the compute instance. Defaults to the API's default image. Find image ids with the `contabo_image` data source, the [API](https://api.contabo.com/#tag/Images/operation/retrieveImage) or `cntb get images`.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Region of the instance, for example `EU`, `UK`, `US-central`, `US-east`, `US-west`, `SIN`, `AUS`, `JPN` or `IND`. Defaults to `EU`. Cannot be changed after creation.",
			},
			"region_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the region.",
			},
			"data_center": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data center of the instance.",
			},
			"product_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Product of the instance, for example `V153`. See the [API documentation](https://api.contabo.com/#tag/Instances/operation/createInstance) for current products. Cannot be changed through the API after creation: upgrade in the customer control panel, then set the new product id.",
			},
			"product_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the product.",
			},
			"ip_config": instanceIpConfigSchema(),
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Mac address of the instance.",
			},
			"ram_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Image ram size in megabyte.",
			},
			"cpu_cores": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU core count of the instance.",
			},
			"disk_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Image disk size of the instance in megabyte.",
			},
			"os_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of operating system (OS) installed on the instance.",
			},
			"ssh_keys": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "CAUTION: On updating this value your server will be reinstalled! Array of `secretIds` of public SSH keys for logging into as defaultUser with administrator/root privileges. Applies to Linux/BSD systems. Please refer to Secrets Management API.",
			},
			"root_password": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "CAUTION: On updating this value your server will be reinstalled! Secret id of the root password of the compute instance.",
			},
			"created_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation date of the compute instance.",
			},
			"cancel_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The date on which the instance will be cancelled. Read only in practice: setting it has no effect.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the compute instance, for example `provisioning`, `installing`, `running`, `stopped`, `error` or `uninstalled`.",
			},
			"v_host_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Identifier of the host system.",
			},
			"add_ons": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Description: "Add-ons of the instance. On create, every listed add-on is ordered with the instance. On existing instances, adding private networking (`1477`) or automated backup (`2214`) orders it through the upgrade API; other add-ons can only be ordered with the instance or in the customer control panel, and removing add-ons has no effect. When set, the state lists only the configured add-ons; the `contabo_instance` data source lists all of them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Id of the Addon. Please refer to list [here](https://contabo.com/en/product-list/?show_ids=true).",
						},
						"quantity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Optional:    true,
							Description: "The number of Addons you wish to aquire.",
						},
					},
				},
			},
			"error_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If the instance is in an error state (see status property), the error message can be seen in this field.",
			},
			"product_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance's category depending on Product Id, for example `ssd`, `nvme` or `vds`.",
			},
			"user_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CAUTION: On updating this value your server will be reinstalled! Cloud-Init Config in order to customize during start of compute instance. The API does not return it, so changes made outside Terraform are not detected.",
			},
			"license": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Additional license in order to enhance your chosen product. It is mainly needed for software licenses on your product (not needed for windows). See our [api documentation](https://api.contabo.com/#tag/Instances/operation/createInstance) for all available licenses. Only used on create.",
			},
			"default_user": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"root", "admin", "administrator"}, false),
				Description:  "CAUTION: On updating this value your server will be reinstalled! Default user name created for login during (re-)installation with administrative privileges. Allowed values for Linux/BSD are admin (use sudo to apply administrative privileges like root) or root. Allowed values for Windows are admin (has administrative privileges like administrator) or administrator.",
			},
			"application_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "CAUTION: On updating this value your server will be reinstalled! Id of an application to install with the image.",
			},
			"period": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Initial contract period in months, for example 1, 3, 6 or 12. The default setting is 1 month. Only used on create.",
			},
			"additional_ips": instanceAdditionalIpsSchema(),
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)

	if existingId := d.Get("existing_instance_id").(string); existingId != "" {
		return adoptInstance(ctx, d, m, existingId)
	}

	request := map[string]interface{}{}
	setIfNotEmpty(request, "displayName", d.Get("display_name").(string))
	setIfNotEmpty(request, "imageId", d.Get("image_id").(string))
	setIfNotEmpty(request, "region", d.Get("region").(string))
	setIfNotEmpty(request, "productId", d.Get("product_id").(string))
	setIfNotEmpty(request, "userData", d.Get("user_data").(string))
	setIfNotEmpty(request, "license", d.Get("license").(string))
	setIfNotEmpty(request, "defaultUser", d.Get("default_user").(string))
	setIfNotEmpty(request, "applicationId", d.Get("application_id").(string))
	if sshKeys := interfacesToInt64s(d.Get("ssh_keys").([]interface{})); len(sshKeys) > 0 {
		request["sshKeys"] = sshKeys
	}
	if rootPassword := d.Get("root_password").(int); rootPassword > 0 {
		request["rootPassword"] = rootPassword
	}
	period := d.Get("period").(int)
	if period == 0 {
		period = 1
	}
	request["period"] = period
	if addOns := expandInstanceAddOns(d.Get("add_ons").([]interface{})); len(addOns) > 0 {
		request["addOns"] = map[string]interface{}{"addonsIds": addOns}
	}

	startedAt := time.Now()
	var res struct {
		Data []struct {
			InstanceId int64 `json:"instanceId"`
		} `json:"data"`
	}
	httpResp, err := apiRequest(ctx, client, http.MethodPost, "/v1/compute/instances", nil, request, &res)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	if len(res.Data) != 1 {
		return MultipleDataObjectsError(nil)
	}
	instanceId := res.Data[0].InstanceId
	// Save the id before waiting: if waiting fails, the paid instance stays
	// in state (as tainted) instead of being ordered again on the next apply.
	d.SetId(strconv.FormatInt(instanceId, 10))

	_, diags := waitForInstanceReady(ctx, client, instanceId, d.Timeout(schema.TimeoutCreate))
	if diags.HasError() {
		return diags
	}
	diags = append(diags, duplicateInstanceWarnings(ctx, client, instanceId, d.Get("display_name").(string), startedAt)...)
	return append(diags, resourceInstanceRead(ctx, d, m)...)
}

// adoptInstance manages an existing instance without reinstalling it.
func adoptInstance(ctx context.Context, d *schema.ResourceData, m interface{}, existingId string) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(existingId, 10, 64)
	if err != nil {
		return diag.Errorf("existing_instance_id must be a numeric instance id: %s", err)
	}
	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	d.SetId(existingId)

	if displayName := d.Get("display_name").(string); displayName != "" && displayName != instance.DisplayName {
		if diags := patchInstanceDisplayName(ctx, client, instanceId, displayName); diags.HasError() {
			return diags
		}
	}
	if diags := upgradeInstanceAddOns(ctx, client, instanceId, d.Get("add_ons").([]interface{})); diags.HasError() {
		return diags
	}
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)

	instanceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("invalid instance id %q: %s", d.Id(), err)
	}

	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		if removeFromStateIfNotFound(d, httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}

	values := flattenInstance(*instance)
	values["id"] = d.Id()
	values["add_ons"] = filterAddOnsToPrior(instance.AddOns, d.Get("add_ons").([]interface{}))
	if instance.ApplicationId != nil && *instance.ApplicationId != "" {
		values["application_id"] = *instance.ApplicationId
	}
	return setSchemaValues(d, resourceInstance().Schema, values)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("invalid instance id %q: %s", d.Id(), err)
	}

	var diags diag.Diagnostics

	if d.HasChange("display_name") {
		if displayName := d.Get("display_name").(string); displayName != "" {
			if diags := patchInstanceDisplayName(ctx, client, instanceId, displayName); diags.HasError() {
				return diags
			}
		}
	}

	if shouldReinstall(d) {
		if diags := reinstallInstance(ctx, client, instanceId, reinstallRequestFromResource(d), d.Timeout(schema.TimeoutUpdate)); diags.HasError() {
			return diags
		}
	}

	if d.HasChange("add_ons") {
		oldRaw, newRaw := d.GetChange("add_ons")
		if removed := addOnIdsOnlyIn(oldRaw.([]interface{}), newRaw.([]interface{})); len(removed) > 0 {
			diags = append(diags, warningDiag(
				"Add-ons are not removed",
				fmt.Sprintf("The API cannot remove add-ons from an instance. Add-ons %v stay on instance %d; cancel them in the customer control panel.", removed, instanceId),
			))
		}
		if upgradeDiags := upgradeInstanceAddOns(ctx, client, instanceId, newRaw.([]interface{})); upgradeDiags.HasError() {
			return append(diags, upgradeDiags...)
		}
	}

	_ = d.Set("last_updated", time.Now().Format(time.RFC850))
	return append(diags, resourceInstanceRead(ctx, d, m)...)
}

// shouldReinstall reports whether an attribute that requires a reinstall has
// changed to a value. Removing user_data or root_password from the
// configuration does not reinstall the instance.
func shouldReinstall(d *schema.ResourceData) bool {
	for _, attribute := range instanceReinstallAttributes {
		if !d.HasChange(attribute) {
			continue
		}
		switch value := d.Get(attribute).(type) {
		case string:
			if value != "" {
				return true
			}
		case int:
			if value > 0 {
				return true
			}
		case []interface{}:
			if len(value) > 0 {
				return true
			}
		}
	}
	return false
}

// reinstallRequestFromResource builds a reinstall request from every
// reinstall attribute, changed or not: the reinstall replaces the whole
// installation, so an attribute left out would be lost (issue #49: user_data
// was dropped when only image_id changed; issue #45: default_user was reset
// to admin).
func reinstallRequestFromResource(d *schema.ResourceData) map[string]interface{} {
	request := map[string]interface{}{
		"imageId": d.Get("image_id").(string),
	}
	setIfNotEmpty(request, "userData", d.Get("user_data").(string))
	setIfNotEmpty(request, "defaultUser", d.Get("default_user").(string))
	setIfNotEmpty(request, "applicationId", d.Get("application_id").(string))
	if sshKeys := interfacesToInt64s(d.Get("ssh_keys").([]interface{})); len(sshKeys) > 0 {
		request["sshKeys"] = sshKeys
	}
	if rootPassword := d.Get("root_password").(int); rootPassword > 0 {
		request["rootPassword"] = rootPassword
	}
	return request
}

func reinstallInstance(
	ctx context.Context,
	client *openapi.APIClient,
	instanceId int64,
	request map[string]interface{},
	timeout time.Duration,
) diag.Diagnostics {
	if imageId, _ := request["imageId"].(string); imageId == "" {
		instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
		if err != nil {
			return apiErrorDiags(httpResp, err)
		}
		if instance.ImageId == nil || *instance.ImageId == "" {
			return diag.Errorf("cannot reinstall instance %d: image_id is not set and the instance reports no image", instanceId)
		}
		request["imageId"] = *instance.ImageId
	}
	httpResp, err := apiRequest(ctx, client, http.MethodPut, instancePath(instanceId), nil, request, nil)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	_, diags := waitForInstanceReady(ctx, client, instanceId, timeout)
	return diags
}

func patchInstanceDisplayName(ctx context.Context, client *openapi.APIClient, instanceId int64, displayName string) diag.Diagnostics {
	httpResp, err := apiRequest(ctx, client, http.MethodPatch, instancePath(instanceId), nil,
		map[string]interface{}{"displayName": displayName}, nil)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	return nil
}

// upgradeInstanceAddOns orders the configured add-ons that the instance does
// not have yet, as far as the upgrade API offers them.
func upgradeInstanceAddOns(ctx context.Context, client *openapi.APIClient, instanceId int64, configured []interface{}) diag.Diagnostics {
	wanted := addOnIds(configured)
	if len(wanted) == 0 {
		return nil
	}
	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	missing := missingAddOnIds(wanted, instance.AddOns)
	if len(missing) == 0 {
		return nil
	}
	request := map[string]interface{}{}
	var unsupported []int64
	for _, id := range missing {
		if key, ok := instanceUpgradeAddOns[id]; ok {
			request[key] = map[string]interface{}{}
		} else {
			unsupported = append(unsupported, id)
		}
	}
	if len(unsupported) > 0 {
		return diag.Errorf("instance %d does not have add-ons %v and the API can only add private networking (1477) and backup (2214) to an existing instance. Order the add-ons in the customer control panel, or remove them from add_ons.", instanceId, unsupported)
	}
	httpResp, err = apiRequest(ctx, client, http.MethodPost, instancePath(instanceId)+"/upgrade", nil, request, nil)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	instanceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("invalid instance id %q: %s", d.Id(), err)
	}

	if !d.Get("cancel_on_destroy").(bool) {
		return diag.Diagnostics{warningDiag(
			"Instance not cancelled",
			fmt.Sprintf("cancel_on_destroy is false, so instance %d was only removed from the Terraform state. It keeps running and billing.", instanceId),
		)}
	}

	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		if isNotFound(httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}
	if instance.cancelDate() != "" {
		return nil
	}

	httpResp, err = apiRequest(ctx, client, http.MethodPost, instancePath(instanceId)+"/cancel", nil, map[string]interface{}{}, nil)
	if err != nil {
		if isNotFound(httpResp) {
			return nil
		}
		diags := apiErrorDiags(httpResp, err)
		if httpResp != nil && httpResp.StatusCode >= 400 && httpResp.StatusCode < 500 {
			diags[0].Detail += "\n\nThe API refused to cancel the instance. Instances ordered in the customer control panel, or with a contract that cannot be cancelled yet, may not be cancellable through the API. Set cancel_on_destroy = false to remove the instance from the Terraform state without cancelling it."
		}
		return diags
	}
	return nil
}

// resourceInstanceCustomizeDiff rejects changes the API cannot make, at plan
// time instead of silently ignoring them during apply.
func resourceInstanceCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	if d.Id() == "" {
		return nil
	}
	for _, attribute := range []string{"product_id", "region"} {
		if !d.HasChange(attribute) {
			continue
		}
		oldValue, newValue := d.GetChange(attribute)
		if oldValue.(string) == "" || newValue.(string) == "" {
			continue
		}
		return fmt.Errorf(
			"%s cannot be changed from %q to %q: the Contabo API cannot change it on an existing instance, and replacing the instance would cancel it. Change it in the customer control panel and then update the configuration, or create a new instance",
			attribute, oldValue, newValue,
		)
	}

	if d.HasChange("add_ons") && m != nil {
		_, newRaw := d.GetChange("add_ons")
		wanted := addOnIds(newRaw.([]interface{}))
		var unsupported []int64
		for _, id := range wanted {
			if _, ok := instanceUpgradeAddOns[id]; !ok {
				unsupported = append(unsupported, id)
			}
		}
		if len(unsupported) == 0 {
			return nil
		}
		client, ok := m.(*openapi.APIClient)
		if !ok {
			return nil
		}
		instanceId, err := strconv.ParseInt(d.Id(), 10, 64)
		if err != nil {
			return nil
		}
		instance, _, err := retrieveInstance(ctx, client, instanceId)
		if err != nil {
			// Read reports API errors; do not fail the plan here.
			return nil
		}
		if missing := missingAddOnIds(unsupported, instance.AddOns); len(missing) > 0 {
			return fmt.Errorf(
				"add-ons %v cannot be added to existing instance %d: the API can only add private networking (1477) and backup (2214). Order them in the customer control panel or remove them from add_ons",
				missing, instanceId,
			)
		}
	}
	return nil
}

// duplicateInstanceWarnings warns about other instances with the same display
// name that were created during this create call. Issue #36 reports several
// instances ordered by one apply; the provider sends a single order, so this
// only makes such a case visible.
func duplicateInstanceWarnings(ctx context.Context, client *openapi.APIClient, instanceId int64, displayName string, startedAt time.Time) diag.Diagnostics {
	if displayName == "" {
		return nil
	}
	instances, _, err := listInstances(ctx, client, url.Values{"displayName": []string{displayName}})
	if err != nil {
		return nil
	}
	var duplicates []string
	for _, instance := range instances {
		if instance.InstanceId == instanceId || instance.DisplayName != displayName {
			continue
		}
		if instance.CreatedDate.After(startedAt.Add(-time.Minute)) {
			duplicates = append(duplicates, strconv.FormatInt(instance.InstanceId, 10))
		}
	}
	if len(duplicates) == 0 {
		return nil
	}
	return diag.Diagnostics{warningDiag(
		"Other instances with the same display name were just created",
		fmt.Sprintf("While instance %d was created, instances %s with display name %q were created too. If they are unexpected, cancel them to avoid charges.", instanceId, strings.Join(duplicates, ", "), displayName),
	)}
}

func setIfNotEmpty(request map[string]interface{}, key string, value string) {
	if value != "" {
		request[key] = value
	}
}

func expandInstanceAddOns(configured []interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, item := range configured {
		addOn, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, err := strconv.ParseInt(mapString(addOn, "id"), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		quantity, _ := addOn["quantity"].(int)
		if quantity == 0 {
			quantity = 1
		}
		out = append(out, map[string]interface{}{"id": id, "quantity": quantity})
	}
	return out
}

func addOnIds(configured []interface{}) []int64 {
	var ids []int64
	for _, addOn := range expandInstanceAddOns(configured) {
		ids = append(ids, addOn["id"].(int64))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func missingAddOnIds(wanted []int64, present []instanceAddOn) []int64 {
	have := map[int64]bool{}
	for _, addOn := range present {
		have[addOn.Id] = true
	}
	var missing []int64
	for _, id := range wanted {
		if !have[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

// addOnIdsOnlyIn returns the ids in a that are not in b.
func addOnIdsOnlyIn(a []interface{}, b []interface{}) []int64 {
	inB := map[int64]bool{}
	for _, id := range addOnIds(b) {
		inB[id] = true
	}
	var out []int64
	for _, id := range addOnIds(a) {
		if !inB[id] {
			out = append(out, id)
		}
	}
	return out
}

// filterAddOnsToPrior keeps the add-ons listed in the prior state, so that
// add-ons the configuration does not mention (the API lists internal ones
// too) do not show up as changes. Without prior add-ons, for example after
// an import, it returns all of them.
func filterAddOnsToPrior(addOns []instanceAddOn, prior []interface{}) []interface{} {
	if len(addOnIds(prior)) == 0 {
		return flattenAddOns(addOns)
	}
	byId := map[int64]instanceAddOn{}
	for _, addOn := range addOns {
		byId[addOn.Id] = addOn
	}
	// Keep the order of the prior state, which follows the configuration.
	var kept []instanceAddOn
	for _, item := range prior {
		addOn, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, err := strconv.ParseInt(mapString(addOn, "id"), 10, 64)
		if err != nil {
			continue
		}
		if present, ok := byId[id]; ok {
			kept = append(kept, present)
			delete(byId, id)
		}
	}
	return flattenAddOns(kept)
}
