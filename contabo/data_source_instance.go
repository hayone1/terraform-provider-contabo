package contabo

import (
	"context"
	"strconv"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// instanceComputedAttributes returns the read-only instance attributes shared
// by the contabo_instance and contabo_instances data sources.
func instanceComputedAttributes() map[string]*schema.Schema {
	computedString := func(description string) *schema.Schema {
		return &schema.Schema{Type: schema.TypeString, Computed: true, Description: description}
	}
	computedInt := func(description string) *schema.Schema {
		return &schema.Schema{Type: schema.TypeInt, Computed: true, Description: description}
	}
	return map[string]*schema.Schema{
		"name":          computedString("Name of the compute instance, for example `vmi123456`."),
		"display_name":  computedString("The instance name chosen by the customer that will be shown in the customer panel."),
		"image_id":      computedString("Id of the image the instance was installed with."),
		"region":        computedString("Region of the instance, for example `EU`."),
		"region_name":   computedString("Name of the region."),
		"data_center":   computedString("Data center of the instance."),
		"product_id":    computedString("Product of the instance, for example `V153`."),
		"product_name":  computedString("Name of the product."),
		"product_type":  computedString("Category of the product, for example `ssd`, `nvme` or `vds`."),
		"ip_config":     instanceIpConfigSchema(),
		"mac_address":   computedString("Mac address of the instance."),
		"ram_mb":        computedInt("RAM size in megabyte."),
		"cpu_cores":     computedInt("CPU core count of the instance."),
		"disk_mb":       computedInt("Disk size of the instance in megabyte."),
		"os_type":       computedString("Type of operating system (OS) installed on the instance."),
		"created_date":  computedString("The creation date of the compute instance."),
		"cancel_date":   computedString("The date on which the instance will be cancelled, empty when it is not cancelled."),
		"status":        computedString("Status of the compute instance, for example `running`, `stopped` or `installing`."),
		"v_host_id":     computedInt("Identifier of the host system."),
		"error_message": computedString("If the instance is in an error state, the error message."),
		"default_user":  computedString("Default user created during the installation."),
		"ssh_keys": {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeInt},
			Description: "Secret ids of the SSH keys installed for the default user.",
		},
		"add_ons": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "All add-ons of the instance.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id":       computedString("Id of the add-on."),
					"quantity": computedInt("Quantity of the add-on."),
				},
			},
		},
		"additional_ips": instanceAdditionalIpsSchema(),
	}
}

func dataSourceInstance() *schema.Resource {
	attributes := instanceComputedAttributes()
	attributes["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The identifier of the compute instance.",
	}
	// Kept for compatibility with earlier versions of this data source. The
	// API does not return these values, so they are always empty.
	for name, description := range map[string]string{
		"last_updated": "Deprecated: always empty.",
		"user_data":    "Deprecated: the API does not return the user data; always empty.",
		"license":      "Deprecated: the API does not return the license; always empty.",
	} {
		attributes[name] = &schema.Schema{Type: schema.TypeString, Computed: true, Description: description}
	}
	attributes["period"] = &schema.Schema{Type: schema.TypeInt, Computed: true, Description: "Deprecated: the API does not return the contract period; always 0."}
	attributes["additional_ips_v4"] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Additional IPv4 addresses of the instance. Same data as `additional_ips`.",
		Elem:        instanceIpSchema(),
	}

	return &schema.Resource{
		Description: "Reads a compute instance by id. Use the `contabo_instances` data source to search instances.",
		ReadContext: dataSourceInstanceRead,
		Schema:      attributes,
	}
}

func dataSourceInstanceRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	client := m.(*openapi.APIClient)

	instanceId, err := strconv.ParseInt(d.Get("id").(string), 10, 64)
	if err != nil {
		return diag.Errorf("id must be a numeric instance id: %s", err)
	}

	instance, httpResp, err := retrieveInstance(ctx, client, instanceId)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}

	d.SetId(strconv.FormatInt(instance.InstanceId, 10))
	return setSchemaValues(d, dataSourceInstance().Schema, flattenInstance(*instance))
}
